package otx

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/vijayvenkatj/cti-miner/pkg/commons"
	"github.com/vijayvenkatj/cti-miner/pkg/resources"
)

const otxTimeLayout = "2006-01-02T15:04:05.999999"

const modifiedSinceStateKey = "otx_modified_since"

type OTXResponse struct {
	Results  []resources.Pulse `json:"results"`
	Count    int               `json:"count"`
	Previous string            `json:"previous"`
	Next     *string           `json:"next"`
}

type PulseStore interface {
	SavePulse(ctx context.Context, pulse resources.Pulse) error
}

type StateStore interface {
	SaveIngestionState(ctx context.Context, key, value string) error
	LoadIngestionState(ctx context.Context, key string) (string, bool, error)
}

type Poller struct {
	BaseURL       string
	ModifiedSince time.Time

	InitialBackoff time.Duration
	MaxBackoff     time.Duration

	Client    *Client
	Publisher *commons.KafkaWriter
	Store     PulseStore
	State     StateStore
}

func NewPoller(baseURL string, modifiedSince time.Time, initialBackoff, maxBackoff time.Duration, client *Client, publisher *commons.KafkaWriter) *Poller {
	return &Poller{
		BaseURL:        baseURL,
		ModifiedSince:  modifiedSince,
		InitialBackoff: initialBackoff,
		MaxBackoff:     maxBackoff,
		Client:         client,
		Publisher:      publisher,
	}
}

func (p *Poller) Run(ctx context.Context) {
	if p.State != nil {
		if value, ok, err := p.State.LoadIngestionState(ctx, modifiedSinceStateKey); err != nil {
			log.Println("error loading ingestion state", err)
		} else if ok {
			if parsed, err := time.Parse(time.RFC3339, value); err == nil {
				p.ModifiedSince = parsed
			}
		}
	}

	delay := p.InitialBackoff

	for {
		err := p.Poll(ctx)
		if err == nil {
			delay = p.InitialBackoff
		}

		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			fmt.Println("context done:", ctx.Err())
			return
		case <-timer.C:
		}

		if err != nil {
			delay *= 2
			delay = min(delay, p.MaxBackoff)
		}
	}
}

func (p *Poller) Poll(ctx context.Context) error {
	u, err := url.Parse(p.BaseURL)
	if err != nil {
		return err
	}

	query := u.Query()
	query.Set("modified_since", p.ModifiedSince.UTC().Format("2006-01-02T15:04:05"))
	query.Set("limit", "10")
	query.Set("page", "1")
	u.RawQuery = query.Encode()

	for {
		resp, err := p.Client.Get(ctx, u.String())
		if err != nil {
			return err
		}

		var response OTXResponse
		if err := json.Unmarshal(resp, &response); err != nil {
			return err
		}

		for _, result := range response.Results {
			modified, err := time.ParseInLocation(otxTimeLayout, result.Modified, time.UTC)
			if err != nil {
				return err
			}

			if modified.After(p.ModifiedSince) {
				p.ModifiedSince = modified
			}

			value, err := json.Marshal(result)
			if err != nil {
				return err
			}
			if pubErr := p.Publisher.WriteMessage(ctx, []byte(result.ID), value); pubErr != nil {
				log.Println("error publishing result", result.ID, pubErr)
			}

			if p.Store != nil {
				if storeErr := p.Store.SavePulse(ctx, result); storeErr != nil {
					log.Println("error storing pulse", result.ID, storeErr)
				}
			}

			if p.State != nil {
				value := p.ModifiedSince.UTC().Format(time.RFC3339)
				if stateErr := p.State.SaveIngestionState(ctx, modifiedSinceStateKey, value); stateErr != nil {
					log.Println("error saving ingestion state", stateErr)
				}
			}
		}

		if response.Next == nil || *response.Next == "" {
			return errors.New("end of stream")
		}

		u, err = url.Parse(*response.Next)
		if err != nil {
			return err
		}
	}
}
