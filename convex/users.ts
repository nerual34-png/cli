import { v } from "convex/values";
import { mutation, query } from "./_generated/server";

// Create a new user
export const create = mutation({
  args: {
    name: v.string(),
    email: v.optional(v.string()),
  },
  handler: async (ctx, { name, email }) => {
    const now = Date.now();
    const userId = await ctx.db.insert("users", {
      name,
      email,
      totalXp: 0,
      weeklyXp: 0,
      level: 1,
      createdAt: now,
      lastActiveAt: now,
    });
    return userId;
  },
});

// Get user by ID
export const get = query({
  args: { userId: v.id("users") },
  handler: async (ctx, { userId }) => {
    return await ctx.db.get(userId);
  },
});

// Get user by email
export const getByEmail = query({
  args: { email: v.string() },
  handler: async (ctx, { email }) => {
    return await ctx.db
      .query("users")
      .withIndex("by_email", (q) => q.eq("email", email))
      .unique();
  },
});

// Update user's group
export const joinGroup = mutation({
  args: {
    userId: v.id("users"),
    groupId: v.id("groups"),
  },
  handler: async (ctx, { userId, groupId }) => {
    const user = await ctx.db.get(userId);
    if (!user) throw new Error("User not found");

    await ctx.db.patch(userId, {
      groupId,
      lastActiveAt: Date.now(),
    });

    // Log activity
    await ctx.db.insert("activity", {
      groupId,
      userId,
      type: "joined_group",
      createdAt: Date.now(),
    });

    return true;
  },
});

// Update user's XP (called after quest completion)
export const addXp = mutation({
  args: {
    userId: v.id("users"),
    xp: v.number(),
  },
  handler: async (ctx, { userId, xp }) => {
    const user = await ctx.db.get(userId);
    if (!user) throw new Error("User not found");

    const newTotalXp = user.totalXp + xp;
    const newWeeklyXp = user.weeklyXp + xp;
    const newLevel = calculateLevel(newTotalXp);
    const leveledUp = newLevel > user.level;

    await ctx.db.patch(userId, {
      totalXp: newTotalXp,
      weeklyXp: newWeeklyXp,
      level: newLevel,
      lastActiveAt: Date.now(),
    });

    return { newTotalXp, newWeeklyXp, newLevel, leveledUp };
  },
});

// Get leaderboard for a group
export const getLeaderboard = query({
  args: {
    groupId: v.id("groups"),
    limit: v.optional(v.number()),
  },
  handler: async (ctx, { groupId, limit = 10 }) => {
    const users = await ctx.db
      .query("users")
      .withIndex("by_group", (q) => q.eq("groupId", groupId))
      .collect();

    // Sort by weekly XP descending
    users.sort((a, b) => b.weeklyXp - a.weeklyXp);

    // Add rank and return top N
    return users.slice(0, limit).map((user, index) => ({
      rank: index + 1,
      userId: user._id,
      userName: user.name,
      level: user.level,
      weeklyXp: user.weeklyXp,
      totalXp: user.totalXp,
    }));
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
