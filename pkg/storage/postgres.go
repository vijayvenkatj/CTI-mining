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
	err := p.qb.Select("id", "name", "description", "author_name", "modified", "created", "revision", "tlp", "public", "adversary").
		From("pulses").
		Where(sq.Eq{"id": id}).
		RunWith(p.db).
		QueryRowContext(ctx).
		Scan(&pulse.ID, &pulse.Name, &pulse.Description, &pulse.AuthorName, &pulse.Modified, &pulse.Created, &pulse.Revision, &pulse.TLP, &pulse.Public, &pulse.Adversary)
	if err == sql.ErrNoRows {
		return resources.Pulse{}, ErrNotFound
	}
	if err != nil {
		return resources.Pulse{}, err
	}

	rows, err := p.qb.Select("id", "pulse_id", "indicator", "type", "created", "title", "is_active").
		From("indicators").
		Where(sq.Eq{"pulse_id": id}).
		RunWith(p.db).
		QueryContext(ctx)
	if err != nil {
		return resources.Pulse{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var indicator resources.Indicator
		if err := rows.Scan(&indicator.ID, &indicator.PulseID, &indicator.Indicator, &indicator.Type, &indicator.Created, &indicator.Title, &indicator.IsActive); err != nil {
			return resources.Pulse{}, err
		}
		pulse.Indicators = append(pulse.Indicators, indicator)
	}
	if err := rows.Err(); err != nil {
		return resources.Pulse{}, err
	}

	return pulse, nil
}

func (p *Postgres) GetIndicator(ctx context.Context, id int64) (resources.Indicator, error) {
	var indicator resources.Indicator
	err := p.qb.Select("id", "pulse_id", "indicator", "type", "created", "title", "is_active").
		From("indicators").
		Where(sq.Eq{"id": id}).
		RunWith(p.db).
		QueryRowContext(ctx).
		Scan(&indicator.ID, &indicator.PulseID, &indicator.Indicator, &indicator.Type, &indicator.Created, &indicator.Title, &indicator.IsActive)
	if err == sql.ErrNoRows {
		return resources.Indicator{}, ErrNotFound
	}
	return indicator, err
}
