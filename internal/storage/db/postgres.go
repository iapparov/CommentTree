package db

import (
	"commentTree/internal/app/domain"
	"commentTree/internal/config"
	"context"
	"database/sql"
	"fmt"
	wbdb "github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/retry"
	wbzlog "github.com/wb-go/wbf/zlog"
	"strings"
)

type Postgres struct {
	db  *wbdb.DB
	cfg *config.RetrysConfig
}

func NewPostgres(cfg *config.AppConfig) (*Postgres, error) {
	masterDSN := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBConfig.Master.Host,
		cfg.DBConfig.Master.Port,
		cfg.DBConfig.Master.User,
		cfg.DBConfig.Master.Password,
		cfg.DBConfig.Master.DBName,
	)

	slaveDSNs := make([]string, 0, len(cfg.DBConfig.Slaves))
	for _, slave := range cfg.DBConfig.Slaves {
		dsn := fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			slave.Host,
			slave.Port,
			slave.User,
			slave.Password,
			slave.DBName,
		)
		slaveDSNs = append(slaveDSNs, dsn)
	}
	var opts wbdb.Options
	opts.ConnMaxLifetime = cfg.DBConfig.ConnMaxLifetime
	opts.MaxIdleConns = cfg.DBConfig.MaxIdleConns
	opts.MaxOpenConns = cfg.DBConfig.MaxOpenConns
	db, err := wbdb.New(masterDSN, slaveDSNs, &opts)
	if err != nil {
		wbzlog.Logger.Debug().Msg("Failed to connect to Postgres")
		return nil, err
	}
	wbzlog.Logger.Info().Msg("Connected to Postgres")
	return &Postgres{db: db, cfg: &cfg.RetrysConfig}, nil
}

func (p *Postgres) Close() error {
	err := p.db.Master.Close()
	if err != nil {
		wbzlog.Logger.Debug().Msg("Failed to close Postgres connection")
		return err
	}
	for _, slave := range p.db.Slaves {
		if slave != nil {
			err := slave.Close()
			if err != nil {
				wbzlog.Logger.Debug().Msg("Failed to close Postgres slave connection")
				return err
			}
		}
	}
	return nil
}

func (p *Postgres) SaveComment(text, parentID string) (*app.Comment, error) {

	comment, err := app.NewComment(parentID, text)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	query := `
		INSERT INTO comments (id, text, createdAt, ParentID, status)
		VALUES($1, $2, $3, $4, 'active')
	`
	_, err = p.db.ExecWithRetry(ctx, retry.Strategy{Attempts: p.cfg.Attempts, Delay: p.cfg.Delay, Backoff: p.cfg.Backoffs}, query,
		comment.ID,
		comment.Text,
		comment.CreatedAt,
		comment.ParentID,
	)
	if err != nil {
		wbzlog.Logger.Error().Err(err).Msg("Failed to execute insert comment query")
		return nil, err
	}
	return comment, nil
}

func (p *Postgres) GetComments(parentId string, sortAsc string, page, pageSize int) ([]app.Comment, error) {
	ctx := context.Background()

	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 50 // значение по умолчанию
	}
	offset := (page - 1) * pageSize

	order := strings.ToUpper(sortAsc)
	if order != "ASC" && order != "DESC" {
		order = "ASC"
	}

	var query string
	var args []interface{}

	if parentId == "" {
		// Все активные комментарии
		query = fmt.Sprintf(`
			SELECT id, text, createdAt, parentId 
			FROM comments 
			WHERE status = 'active'
			ORDER BY createdAt %s
			LIMIT $1 OFFSET $2;
		`, order)
		args = []interface{}{pageSize, offset}
	} else {
		// Рекурсивное дерево с указанным parentId
		query = fmt.Sprintf(`
			WITH RECURSIVE tree AS (
				SELECT * FROM comments WHERE id = $1
				UNION ALL
				SELECT c.*
				FROM comments c
				INNER JOIN tree t ON c.ParentID = t.id
				WHERE c.status = 'active'
			)
			SELECT id, text, createdAt, parentId FROM tree
			ORDER BY createdAt %s
			LIMIT $2 OFFSET $3;
		`, order)
		args = []interface{}{parentId, pageSize, offset}
	}

	rows, err := p.db.QueryWithRetry(
		ctx,
		retry.Strategy{Attempts: p.cfg.Attempts, Delay: p.cfg.Delay, Backoff: p.cfg.Backoffs},
		query,
		args...,
	)
	if err != nil {
		wbzlog.Logger.Error().Err(err).Msg("Failed to execute select comments query")
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			wbzlog.Logger.Error().Err(err).Msg("Failed to close rows")
		}
	}()

	var comments []app.Comment
	for rows.Next() {
		var c app.Comment
		err := rows.Scan(&c.ID, &c.Text, &c.CreatedAt, &c.ParentID)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}
			wbzlog.Logger.Error().Err(err).Msg("Failed to scan comment row")
			return nil, err
		}
		comments = append(comments, c)
	}

	if err = rows.Err(); err != nil {
		wbzlog.Logger.Error().Err(err).Msg("Row iteration error")
		return nil, err
	}
	return comments, nil
}

func (p *Postgres) DeleteComments(id string) error {
	ctx := context.Background()

	query := `
		WITH RECURSIVE tree AS (
			SELECT id FROM comments WHERE id = $1
			UNION ALL
			SELECT c.id
			FROM comments c
			INNER JOIN tree t ON c.ParentID = t.id
		)
		UPDATE comments
		SET status = 'deleted'
		WHERE id IN (SELECT id FROM tree);
	`

	_, err := p.db.ExecWithRetry(ctx, retry.Strategy{Attempts: p.cfg.Attempts, Delay: p.cfg.Delay, Backoff: p.cfg.Backoffs}, query, id)
	if err != nil {
		wbzlog.Logger.Error().Err(err).Msg("Failed to execute insert comment query")
		return err
	}
	return nil
}

func (p *Postgres) SearchComments(text string, sortAsc string, page, pageSize int) ([]app.Comment, error) {
	ctx := context.Background()

	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 50
	}
	offset := (page - 1) * pageSize

	order := strings.ToLower(sortAsc)
	if order != "asc" && order != "desc" {
		order = "ASC"
	}

	query := fmt.Sprintf(`
		SELECT id, text, createdAt, parentId
		FROM comments
		WHERE status = 'active'
		AND to_tsvector('simple', text) @@ plainto_tsquery('simple', $1)
		ORDER BY createdAt %s
		LIMIT $2 OFFSET $3;
	`, order)

	rows, err := p.db.QueryWithRetry(ctx,
		retry.Strategy{Attempts: p.cfg.Attempts, Delay: p.cfg.Delay, Backoff: p.cfg.Backoffs},
		query, text, pageSize, offset)
	if err != nil {
		wbzlog.Logger.Error().Err(err).Msg("Failed to execute search comments query")
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			wbzlog.Logger.Error().Err(err).Msg("Failed to close rows")
		}
	}()

	var comments []app.Comment
	for rows.Next() {
		var c app.Comment
		err := rows.Scan(&c.ID, &c.Text, &c.CreatedAt, &c.ParentID)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}
			wbzlog.Logger.Error().Err(err).Msg("Failed to scan comment row")
			return nil, err
		}
		comments = append(comments, c)
	}

	if err = rows.Err(); err != nil {
		wbzlog.Logger.Error().Err(err).Msg("Row iteration error")
		return nil, err
	}

	return comments, nil
}
