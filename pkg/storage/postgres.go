package storage

import (
	"context"
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/vijayvenkatj/cti-miner/pkg/resources"
)

var ErrNotFound = errors.New("not found")

var (
	pulseColumns     = []string{"id", "name", "description", "author_name", "modified", "created", "revision", "tlp", "public", "adversary"}
	indicatorColumns = []string{"id", "pulse_id", "indicator", "type", "created", "title", "is_active"}
)

type scanner interface {
	Scan(dest ...any) error
}

func scanPulse(s scanner, pulse *resources.Pulse) error {
	return s.Scan(&pulse.ID, &pulse.Name, &pulse.Description, &pulse.AuthorName, &pulse.Modified, &pulse.Created, &pulse.Revision, &pulse.TLP, &pulse.Public, &pulse.Adversary)
}

func scanIndicator(s scanner, indicator *resources.Indicator) error {
	return s.Scan(&indicator.ID, &indicator.PulseID, &indicator.Indicator, &indicator.Type, &indicator.Created, &indicator.Title, &indicator.IsActive)
}

type Postgres struct {
	db *sql.DB
	qb sq.StatementBuilderType
}

func NewPostgres(dsn string) (*Postgres, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Postgres{
		db: db,
		qb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}, nil
}

func (p *Postgres) Close() error {
	return p.db.Close()
}

func (p *Postgres) SavePulse(ctx context.Context, pulse resources.Pulse) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = p.qb.Insert("pulses").
		Columns("id", "name", "description", "author_name", "modified", "created", "revision", "tlp", "public", "adversary").
		Values(pulse.ID, pulse.Name, pulse.Description, pulse.AuthorName, pulse.Modified, pulse.Created, pulse.Revision, pulse.TLP, pulse.Public, pulse.Adversary).
		Suffix("ON CONFLICT (id) DO NOTHING").
		RunWith(tx).
		ExecContext(ctx)
	if err != nil {
		return err
	}

	for _, indicator := range pulse.Indicators {
		_, err = p.qb.Insert("indicators").
			Columns("id", "pulse_id", "indicator", "type", "created", "title", "is_active").
			Values(indicator.ID, pulse.ID, indicator.Indicator, indicator.Type, indicator.Created, indicator.Title, indicator.IsActive).
			Suffix("ON CONFLICT (id) DO NOTHING").
			RunWith(tx).
			ExecContext(ctx)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (p *Postgres) SaveEdge(ctx context.Context, edge resources.Edge) error {
	_, err := p.qb.Insert("edges").
		Columns("source_pulse_id", "target_pulse_id").
		Values(edge.Source, edge.Target).
		Suffix("ON CONFLICT (source_pulse_id, target_pulse_id) DO NOTHING").
		RunWith(p.db).
		ExecContext(ctx)
	return err
}

func (p *Postgres) SaveIngestionState(ctx context.Context, key, value string) error {
	_, err := p.qb.Insert("ingestion_state").
		Columns("id", "value").
		Values(key, value).
		Suffix("ON CONFLICT (id) DO UPDATE SET value = EXCLUDED.value, updated_at = now()").
		RunWith(p.db).
		ExecContext(ctx)
	return err
}

func (p *Postgres) LoadIngestionState(ctx context.Context, key string) (string, bool, error) {
	var value string
	err := p.qb.Select("value").
		From("ingestion_state").
		Where(sq.Eq{"id": key}).
		RunWith(p.db).
		QueryRowContext(ctx).
		Scan(&value)

	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return value, true, nil
}

func (p *Postgres) ListEdges(ctx context.Context) ([]resources.Edge, error) {
	rows, err := p.qb.Select("source_pulse_id", "target_pulse_id").
		From("edges").
		RunWith(p.db).
		QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	edges := []resources.Edge{}
	for rows.Next() {
		var edge resources.Edge
		if err := rows.Scan(&edge.Source, &edge.Target); err != nil {
			return nil, err
		}
		edges = append(edges, edge)
	}
	return edges, rows.Err()
}

func (p *Postgres) GetPulse(ctx context.Context, id string) (resources.Pulse, error) {
	var pulse resources.Pulse
	row := p.qb.Select(pulseColumns...).From("pulses").Where(sq.Eq{"id": id}).RunWith(p.db).QueryRowContext(ctx)
	if err := scanPulse(row, &pulse); err != nil {
		if err == sql.ErrNoRows {
			return resources.Pulse{}, ErrNotFound
		}
		return resources.Pulse{}, err
	}

	indicators, err := p.PulseIndicators(ctx, id)
	if err != nil {
		return resources.Pulse{}, err
	}
	pulse.Indicators = indicators

	return pulse, nil
}

func (p *Postgres) ListPulses(ctx context.Context) ([]resources.Pulse, error) {
	rows, err := p.qb.Select(pulseColumns...).From("pulses").RunWith(p.db).QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pulses := []resources.Pulse{}
	for rows.Next() {
		var pulse resources.Pulse
		if err := scanPulse(rows, &pulse); err != nil {
			return nil, err
		}
		pulses = append(pulses, pulse)
	}
	return pulses, rows.Err()
}

func (p *Postgres) GetIndicator(ctx context.Context, id int64) (resources.Indicator, error) {
	var indicator resources.Indicator
	row := p.qb.Select(indicatorColumns...).From("indicators").Where(sq.Eq{"id": id}).RunWith(p.db).QueryRowContext(ctx)
	if err := scanIndicator(row, &indicator); err != nil {
		if err == sql.ErrNoRows {
			return resources.Indicator{}, ErrNotFound
		}
		return resources.Indicator{}, err
	}
	return indicator, nil
}

func (p *Postgres) ListIndicators(ctx context.Context) ([]resources.Indicator, error) {
	rows, err := p.qb.Select(indicatorColumns...).From("indicators").RunWith(p.db).QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	indicators := []resources.Indicator{}
	for rows.Next() {
		var indicator resources.Indicator
		if err := scanIndicator(rows, &indicator); err != nil {
			return nil, err
		}
		indicators = append(indicators, indicator)
	}
	return indicators, rows.Err()
}

func (p *Postgres) PulseIndicators(ctx context.Context, pulseID string) ([]resources.Indicator, error) {
	rows, err := p.qb.Select(indicatorColumns...).From("indicators").Where(sq.Eq{"pulse_id": pulseID}).RunWith(p.db).QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	indicators := []resources.Indicator{}
	for rows.Next() {
		var indicator resources.Indicator
		if err := scanIndicator(rows, &indicator); err != nil {
			return nil, err
		}
		indicators = append(indicators, indicator)
	}
	return indicators, rows.Err()
}
