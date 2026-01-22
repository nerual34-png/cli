import { defineSchema, defineTable } from "convex/server";
import { v } from "convex/values";

export default defineSchema({
  users: defineTable({
    name: v.string(),
    email: v.optional(v.string()),
    groupId: v.optional(v.id("groups")),
    totalXp: v.number(),
    weeklyXp: v.number(),
    level: v.number(),
    createdAt: v.number(),
    lastActiveAt: v.number(),
  })
    .index("by_email", ["email"])
    .index("by_group", ["groupId"])
    .index("by_total_xp", ["totalXp"])
    .index("by_weekly_xp", ["groupId", "weeklyXp"]),

  groups: defineTable({
    name: v.string(),
    inviteCode: v.string(),
    createdBy: v.id("users"),
    createdAt: v.number(),
  }).index("by_invite_code", ["inviteCode"]),

  quests: defineTable({
    userId: v.id("users"),
    groupId: v.optional(v.id("groups")),
    title: v.string(),
    xp: v.number(),
    aiReasoning: v.string(),
    status: v.union(v.literal("pending"), v.literal("in_progress"), v.literal("completed")),
    createdAt: v.number(),
    completedAt: v.optional(v.number()),
  })
    .index("by_user", ["userId"])
    .index("by_user_status", ["userId", "status"])
    .index("by_user_created", ["userId", "createdAt"])
    .index("by_group", ["groupId"]),

  activity: defineTable({
    groupId: v.id("groups"),
    userId: v.id("users"),
    type: v.union(
      v.literal("quest_created"),
      v.literal("quest_started"),
      v.literal("quest_completed"),
      v.literal("level_up"),
      v.literal("joined_group")
    ),
    questTitle: v.optional(v.string()),
    xp: v.optional(v.number()),
    newLevel: v.optional(v.number()),
    createdAt: v.number(),
  })
    .index("by_group", ["groupId"])
    .index("by_group_created", ["groupId", "createdAt"]),
});
