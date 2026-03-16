package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ReferralReward holds the schema definition for the ReferralReward entity.
type ReferralReward struct {
	ent.Schema
}

func (ReferralReward) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "referral_rewards"},
	}
}

func (ReferralReward) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("inviter_id"),
		field.Int64("invitee_id"),
		field.Int64("trigger_redeem_code_id").
			Optional().
			Nillable(),
		field.Float("reward_amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}),
		field.String("reward_type").
			MaxLen(20).
			Default("percentage"),
		field.Float("reward_rate").
			SchemaType(map[string]string{dialect.Postgres: "decimal(10,4)"}).
			Optional().
			Nillable(),
		field.Float("trigger_code_value").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}),
		field.String("status").
			MaxLen(20).
			Default("completed"),
		field.String("notes").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.Time("created_at").
			Immutable().
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (ReferralReward) Edges() []ent.Edge {
	return nil
}

func (ReferralReward) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("inviter_id"),
		index.Fields("invitee_id"),
		index.Fields("created_at"),
	}
}
