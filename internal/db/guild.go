package db

import (
	"context"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GuildUser struct {
	Email    string
	TeamSize int
	Progress float64
}

type Guild struct {
	Name  string
	ID    uuid.UUID
	Users []GuildUser
}

var ErrNoGuild error = fmt.Errorf("No guild found")

// TODO : get progress (percentage of progression on cards)
func GetGuild(dbPool *pgxpool.Pool, user User) ([]Guild, error) {
	query :=
		`WITH guild AS (
			SELECT guilds.id, guilds.name
			FROM guilds
			JOIN user_guilds ON guilds.id = user_guilds.guild_id
			WHERE user_guilds.user_id = @user_id
		)
		SELECT
			users.team_size,
			users.email,
			ROUND(
				(
					SELECT COUNT(*)
					FROM progress
					WHERE progress.user_id = users.id
						AND progress.box_index < users.team_size
						AND progress.done = TRUE
				) * 100.0 / (SELECT COUNT(*) * users.team_size FROM cards),
				2
			),
			guild.name,
			guild.id
		FROM guild
		JOIN user_guilds ON user_guilds.guild_id = guild.id
		JOIN users ON users.id = user_guilds.user_id
		;`
	args := pgx.NamedArgs{"user_id": user.ID}
	rows, err := dbPool.Query(context.Background(), query, args)
	if err != nil {
		return nil, err
	}
	guilds := make(map[uuid.UUID]Guild, 0)
	for rows.Next() {
		var user GuildUser
		var guildName string
		var guildID uuid.UUID
		err = rows.Scan(
			&user.TeamSize,
			&user.Email,
			&user.Progress,
			&guildName,
			&guildID,
		)
		if err != nil {
			return nil, err
		}
		if _, ok := guilds[guildID]; !ok {
			guilds[guildID] = Guild{ID: guildID,
				Name:  guildName,
				Users: []GuildUser{},
			}
		}
		guild := guilds[guildID]
		guild.Users = append(guild.Users, user)
		guilds[guildID] = guild // we must reassign because previous create a copy
	}
	if len(guilds) == 0 {
		return nil, ErrNoGuild
	}
	guildsSlice := make([]Guild, 0, len(guilds))
	for _, v := range guilds {
		guildsSlice = append(guildsSlice, v)
	}
	sort.Slice(guildsSlice, func(i, j int) bool {
		return len(guildsSlice[i].Users) > len(guildsSlice[j].Users)
	})
	return guildsSlice, nil
}

type GuildSize struct {
	ID       uuid.UUID
	Name     string
	Size     int
	IsMember bool
}

// Returns the name and the size of guilds matching the substring
func SearchGuilds(dbPool *pgxpool.Pool, user User, substring string) ([]GuildSize, error) {
	if substring == "" {
		return []GuildSize{}, nil
	}
	query :=
		`SELECT
			guilds.id,
			guilds.name,
			COUNT(user_guilds.user_id),
	        EXISTS (
	            SELECT 1
	            FROM user_guilds
	            WHERE user_guilds.guild_id = guilds.id
	            AND user_guilds.user_id = @user_id
	        ) AS is_member
		FROM guilds
		JOIN user_guilds ON user_guilds.guild_id = guilds.id
		WHERE guilds.name ILIKE @substring
		GROUP BY guilds.id`
	args := pgx.NamedArgs{
		"substring": fmt.Sprintf("%%%s%%", substring),
		"user_id":   user.ID,
	}
	rows, err := dbPool.Query(context.Background(), query, args)
	if err != nil {
		return []GuildSize{}, err
	}
	guildsInfo := make([]GuildSize, 0)
	for rows.Next() {
		var guildInfo GuildSize
		err = rows.Scan(
			&guildInfo.ID,
			&guildInfo.Name,
			&guildInfo.Size,
			&guildInfo.IsMember,
		)
		if err != nil {
			return []GuildSize{}, err
		}
		guildsInfo = append(guildsInfo, guildInfo)
	}
	return guildsInfo, err
}

type DB interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}

func CreateGuild(dbPool DB, ctx context.Context, name string) (uuid.UUID, error) {
	query :=
		`INSERT INTO guilds(id, name)
		VALUES (@id, @name);`
	id := uuid.New()
	args := pgx.NamedArgs{"id": id, "name": name}
	_, err := dbPool.Exec(ctx, query, args)
	return id, err
}
func JoinGuild(dbPool DB, ctx context.Context, guildID uuid.UUID, user User) error {
	query :=
		`INSERT INTO user_guilds(user_id, guild_id)
		VALUES (@user_id, @guild_id);`
	args := pgx.NamedArgs{"user_id": user.ID, "guild_id": guildID}
	_, err := dbPool.Exec(ctx, query, args)
	return err
}

func LeaveGuild(dbPool DB, ctx context.Context, guildID uuid.UUID, user User) error {
	query :=
		`DELETE FROM user_guilds WHERE user_id = @user_id AND guild_id = @guild_id;`
	args := pgx.NamedArgs{"user_id": user.ID, "guild_id": guildID}
	_, err := dbPool.Exec(ctx, query, args)
	return err
}

func DeleteGuildIfEmpty(dbPool DB, ctx context.Context, guildID uuid.UUID) error {
	query :=
		`DELETE FROM guilds
		WHERE guilds.id = @guild_id
			AND NOT EXISTS (
				SELECT 1
				FROM user_guilds
				WHERE user_guilds.guild_id = @guild_id
			);`
	args := pgx.NamedArgs{"guild_id": guildID}
	_, err := dbPool.Exec(ctx, query, args)
	return err
}
