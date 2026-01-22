import { v } from "convex/values";
import { query } from "./_generated/server";

// Get recent activity for a group
export const getRecent = query({
  args: {
    groupId: v.id("groups"),
    limit: v.optional(v.number()),
  },
  handler: async (ctx, { groupId, limit = 20 }) => {
    const activities = await ctx.db
      .query("activity")
      .withIndex("by_group_created", (q) => q.eq("groupId", groupId))
      .order("desc")
      .take(limit);

    // Fetch user names for each activity
    const activitiesWithNames = await Promise.all(
      activities.map(async (activity) => {
        const user = await ctx.db.get(activity.userId);
        return {
          ...activity,
          userName: user?.name ?? "Unknown",
        };
      })
    );

    return activitiesWithNames;
  },
});

// Get user's recent activity
export const getUserActivity = query({
  args: {
    userId: v.id("users"),
    limit: v.optional(v.number()),
  },
  handler: async (ctx, { userId, limit = 20 }) => {
    const user = await ctx.db.get(userId);
    if (!user || !user.groupId) {
      return [];
    }

    const groupId = user.groupId;
    const activities = await ctx.db
      .query("activity")
      .withIndex("by_group_created", (q) => q.eq("groupId", groupId))
      .order("desc")
      .take(limit);

    // Fetch user names for each activity
    const activitiesWithNames = await Promise.all(
      activities.map(async (activity) => {
        const activityUser = await ctx.db.get(activity.userId);
        return {
          ...activity,
          userName: activityUser?.name ?? "Unknown",
        };
      })
    );

    return activitiesWithNames;
  },
});

// Get activity count for today
export const getTodayCount = query({
  args: { groupId: v.id("groups") },
  handler: async (ctx, { groupId }) => {
    const startOfDay = new Date();
    startOfDay.setHours(0, 0, 0, 0);
    const startTimestamp = startOfDay.getTime();

    const activities = await ctx.db
      .query("activity")
      .withIndex("by_group_created", (q) => q.eq("groupId", groupId))
      .filter((q) => q.gte(q.field("createdAt"), startTimestamp))
      .collect();

    return {
      total: activities.length,
      questsCreated: activities.filter((a) => a.type === "quest_created").length,
      questsCompleted: activities.filter((a) => a.type === "quest_completed").length,
      levelUps: activities.filter((a) => a.type === "level_up").length,
    };
  },
});
