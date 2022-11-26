package postgres

import (
	"context"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/ssssargsian/uniplay/internal/domain"
	"github.com/ssssargsian/uniplay/internal/dto"
	"github.com/ysomad/pgxatomic"
	"go.uber.org/zap"
)

type metricRepo struct {
	log     *zap.Logger
	pool    pgxatomic.Pool
	builder sq.StatementBuilderType
}

func NewMetricRepo(l *zap.Logger, p pgxatomic.Pool, b sq.StatementBuilderType) *metricRepo {
	return &metricRepo{
		log:     l,
		pool:    p,
		builder: b,
	}
}

func (r *metricRepo) GetWeaponMetrics(ctx context.Context, steamID uint64, f domain.WeaponStatsFilter) ([]dto.WeaponMetricSum, error) {
	sb := r.builder.
		Select("weapon_name, metric, SUM(value)").
		From("weapon_metric").
		Where(sq.Eq{"player_steam_id": steamID})

	switch {
	case f.WeaponName != "":
		sb = sb.Where(sq.Eq{"weapon_name": strings.ToLower(f.WeaponName)})
	case f.WeaponClass.Valid():
		sb = sb.Where(sq.Eq{"weapon_class": f.WeaponClass})
	}

	sql, args, err := sb.
		GroupBy("metric, weapon_name").
		OrderBy("weapon_name").
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	m, err := pgx.CollectRows(rows, pgx.RowToStructByPos[dto.WeaponMetricSum])
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (r *metricRepo) GetWeaponClassMetrics(ctx context.Context, steamID uint64, c domain.WeaponClassID) ([]dto.WeaponClassMetricSum, error) {
	sb := r.builder.
		Select("weapon_class, metric, SUM(value)").
		From("weapon_metric").
		Where(sq.Eq{"player_steam_id": steamID})

	if c.Valid() {
		sb = sb.Where(sq.Eq{"weapon_class": c})
	}

	sql, args, err := sb.
		GroupBy("metric, weapon_class").
		OrderBy("weapon_class").
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	m, err := pgx.CollectRows(rows, pgx.RowToStructByPos[dto.WeaponClassMetricSum])
	if err != nil {
		return nil, err
	}

	return m, nil
}
