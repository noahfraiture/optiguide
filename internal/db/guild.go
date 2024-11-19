package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GuildUser struct {
	Email    string
	TeamSize int
	Progress float32
}

type GuildMembers struct {
	Name  string
	ID    uuid.UUID
	Users []GuildUser
}

var ErrNoGuild error = fmt.Errorf("No guild found")

func GetGuild(dbPool *pgxpool.Pool, userID string) (GuildMembers, error) {
	query :=
		`WITH guild AS (
			SELECT guilds.id, guilds.name
			FROM guilds
			JOIN user_guilds ON guilds.id = user_guilds.guild_id
			WHERE user_guilds.user_id = @user_id
		)
		SELECT users.team_size, users.email, guild.name, guild.id
		FROM users
		JOIN user_guilds ON user_guilds.user_id = users.id
		JOIN guild ON guild.id = user_guilds.guild_id;`
	args := pgx.NamedArgs{"user_id": userID}
	rows, err := dbPool.Query(context.Background(), query, args)
	if err != nil {
		return GuildMembers{}, err
	}
	users := make([]GuildUser, 0)
	var guildName string
	var guildID uuid.UUID
	for rows.Next() {
		var user GuildUser
		err = rows.Scan(
			&user.TeamSize,
			&user.Email,
			&guildName,
			&guildID,
		)
		if err != nil {
			return GuildMembers{}, err
		}
		users = append(users, user)
	}
	if len(users) == 0 {
		return GuildMembers{}, ErrNoGuild
	}
	return GuildMembers{Name: guildName, ID: guildID, Users: users}, nil
}

type GuildSize struct {
	ID   uuid.UUID
	Name string
	Size int
}

// Returns the name and the size of guilds matching the substring
func SearchGuilds(dbPool *pgxpool.Pool, substring string) ([]GuildSize, error) {
	query :=
		`SELECT guilds.id, guilds.name, COUNT(user_guilds.user_id)
		FROM guilds
		JOIN user_guilds ON user_guilds.guild_id = guilds.id
		WHERE guilds.name ILIKE @substring
		GROUP BY guilds.id`
	args := pgx.NamedArgs{"substring": fmt.Sprintf("%%%s%%", substring)}
	rows, err := dbPool.Query(context.Background(), query, args)
	if err != nil {
		return []GuildSize{}, err
	}
	guildsInfo := make([]GuildSize, 0)
	for rows.Next() {
		var guildInfo GuildSize
		err = rows.Scan(&guildInfo.ID, &guildInfo.Name, &guildInfo.Size)
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
func JoinGuild(dbPool DB, ctx context.Context, guildID uuid.UUID, userID string) error {
	query :=
		`INSERT INTO user_guilds(user_id, guild_id)
		VALUES (@user_id, @guild_id);`
	args := pgx.NamedArgs{"user_id": userID, "guild_id": guildID}
	_, err := dbPool.Exec(ctx, query, args)
	return err
}

func LeaveGuild(dbPool DB, ctx context.Context, guildID uuid.UUID, userID string) error {
	query :=
		`DELETE FROM user_guilds WHERE user_id = @user_id AND guild_id = @guild_id;`
	args := pgx.NamedArgs{"user_id": userID, "guild_id": guildID}
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
