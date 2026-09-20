package storage

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/vijayvenkatj/cti-miner/pkg/resources"
)

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
