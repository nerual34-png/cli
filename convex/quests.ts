import { v } from "convex/values";
import { mutation, query, action } from "./_generated/server";
import { api } from "./_generated/api";

// Create a new quest (calls AI for XP evaluation)
export const create = mutation({
  args: {
    userId: v.id("users"),
    title: v.string(),
    xp: v.number(),
    aiReasoning: v.string(),
  },
  handler: async (ctx, { userId, title, xp, aiReasoning }) => {
    const user = await ctx.db.get(userId);
    if (!user) throw new Error("User not found");

    const now = Date.now();
    const questId = await ctx.db.insert("quests", {
      userId,
      groupId: user.groupId,
      title,
      xp,
      aiReasoning,
      status: "pending",
      createdAt: now,
    });

    // Log activity if in a group
    if (user.groupId) {
      await ctx.db.insert("activity", {
        groupId: user.groupId,
        userId,
        type: "quest_created",
        questTitle: title,
        xp,
        createdAt: now,
      });
    }

    return { questId, xp, aiReasoning };
  },
});

// Complete a quest (from pending or in_progress → completed)
export const complete = mutation({
  args: { questId: v.id("quests") },
  handler: async (ctx, { questId }) => {
    const quest = await ctx.db.get(questId);
    if (!quest) throw new Error("Quest not found");
    if (quest.status === "completed") throw new Error("Quest already completed");
    // Allow completing from both pending and in_progress

    const user = await ctx.db.get(quest.userId);
    if (!user) throw new Error("User not found");

    const now = Date.now();

    // Update quest status
    await ctx.db.patch(questId, {
      status: "completed",
      completedAt: now,
    });

    // Update user XP
    const newTotalXp = user.totalXp + quest.xp;
    const newWeeklyXp = user.weeklyXp + quest.xp;
    const newLevel = calculateLevel(newTotalXp);
    const leveledUp = newLevel > user.level;

    await ctx.db.patch(user._id, {
      totalXp: newTotalXp,
      weeklyXp: newWeeklyXp,
      level: newLevel,
      lastActiveAt: now,
    });

    // Log activity if in a group
    if (user.groupId) {
      await ctx.db.insert("activity", {
        groupId: user.groupId,
        userId: user._id,
        type: "quest_completed",
        questTitle: quest.title,
        xp: quest.xp,
        createdAt: now,
      });

      if (leveledUp) {
        await ctx.db.insert("activity", {
          groupId: user.groupId,
          userId: user._id,
          type: "level_up",
          newLevel,
          createdAt: now,
        });
      }
    }

    return {
      xpEarned: quest.xp,
      newTotalXp,
      newWeeklyXp,
      leveledUp,
      newLevel,
    };
  },
});

// Start a quest (pending → in_progress)
export const start = mutation({
  args: { questId: v.id("quests") },
  handler: async (ctx, { questId }) => {
    const quest = await ctx.db.get(questId);
    if (!quest) throw new Error("Quest not found");
    if (quest.status !== "pending") throw new Error("Quest must be pending to start");

    const user = await ctx.db.get(quest.userId);
    if (!user) throw new Error("User not found");

    const now = Date.now();

    // Update quest status
    await ctx.db.patch(questId, {
      status: "in_progress",
    });

    // Log activity if in a group
    if (user.groupId) {
      await ctx.db.insert("activity", {
        groupId: user.groupId,
        userId: user._id,
        type: "quest_started",
        questTitle: quest.title,
        createdAt: now,
      });
    }

    return { questId, status: "in_progress" };
  },
});

// Get user's quests
export const list = query({
  args: {
    userId: v.id("users"),
    status: v.optional(v.union(v.literal("pending"), v.literal("in_progress"), v.literal("completed"))),
  },
  handler: async (ctx, { userId, status }) => {
    let quests;
    if (status) {
      quests = await ctx.db
        .query("quests")
        .withIndex("by_user_status", (q) => q.eq("userId", userId).eq("status", status))
        .collect();
    } else {
      quests = await ctx.db
        .query("quests")
        .withIndex("by_user", (q) => q.eq("userId", userId))
        .collect();
    }

    // Sort by createdAt descending
    return quests.sort((a, b) => b.createdAt - a.createdAt);
  },
});

// Get today's quests
export const listToday = query({
  args: { userId: v.id("users") },
  handler: async (ctx, { userId }) => {
    const startOfDay = new Date();
    startOfDay.setHours(0, 0, 0, 0);
    const startTimestamp = startOfDay.getTime();

    const quests = await ctx.db
      .query("quests")
      .withIndex("by_user_created", (q) => q.eq("userId", userId))
      .filter((q) => q.gte(q.field("createdAt"), startTimestamp))
      .collect();

    return quests.sort((a, b) => a.createdAt - b.createdAt);
  },
});

// Delete a quest
export const remove = mutation({
  args: { questId: v.id("quests") },
  handler: async (ctx, { questId }) => {
    const quest = await ctx.db.get(questId);
    if (!quest) throw new Error("Quest not found");
    if (quest.status === "completed") throw new Error("Cannot delete completed quest");

    await ctx.db.delete(questId);
    return true;
  },
});

// Helper function to calculate level from XP
function calculateLevel(xp: number): number {
  const levels = [
    { level: 1, minXp: 0 },
    { level: 2, minXp: 100 },
    { level: 3, minXp: 300 },
    { level: 4, minXp: 600 },
    { level: 5, minXp: 1000 },
    { level: 6, minXp: 1500 },
    { level: 7, minXp: 2200 },
    { level: 8, minXp: 3000 },
    { level: 9, minXp: 4000 },
    { level: 10, minXp: 5500 },
  ];

  let currentLevel = 1;
  for (const l of levels) {
    if (xp >= l.minXp) {
      currentLevel = l.level;
    } else {
      break;
    }
  }
  return currentLevel;
}
