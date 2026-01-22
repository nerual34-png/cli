import { v } from "convex/values";
import { query, action } from "./_generated/server";
import { api } from "./_generated/api";

// Deep philosophical quotes for grinders
const QUOTES = [
  // Jungian / Shadow work
  "Where your fear is, there is your task.",
  "The cave you fear holds the treasure you seek.",
  "What you resist, persists.",
  "No tree grows to heaven unless its roots reach hell.",

  // Stoic wisdom
  "The obstacle is the way.",
  "You could leave life right now. Let that determine what you do.",
  "Waste no time arguing what a good man should be. Be one.",
  "He who fears death will never do anything worthy of a man.",

  // Nietzsche / Existential
  "He who has a why can bear almost any how.",
  "You must have chaos within to give birth to a dancing star.",
  "Become who you are.",
  "What doesn't kill me makes me stronger.",

  // Paradox / Depth
  "Whoever tries to save his life will lose it.",
  "The wound is where the light enters.",
  "Comfort is the enemy of progress.",
  "To live is to suffer. To survive is to find meaning.",

  // Action
  "Ship the fear. Debug later.",
  "Do it scared. Do it anyway.",
  "The grind reveals, it doesn't conceal.",
];

// Get dashboard stats for a user (public query)
export const getStats = query({
  args: { userId: v.id("users") },
  handler: async (ctx, { userId }) => {
    const user = await ctx.db.get(userId);
    if (!user) {
      return null;
    }

    const startOfDay = new Date();
    startOfDay.setHours(0, 0, 0, 0);
    const todayStart = startOfDay.getTime();

    // Get today's quests for this user
    const todayQuests = await ctx.db
      .query("quests")
      .withIndex("by_user_created", (q) => q.eq("userId", userId))
      .filter((q) => q.gte(q.field("createdAt"), todayStart))
      .collect();

    const todayCompleted = todayQuests.filter((q) => q.status === "completed");
    const todayXP = todayCompleted.reduce((sum, q) => sum + q.xp, 0);

    // Get group stats if user is in a group
    let groupStats = null;
    let memberStats: Array<{
      name: string;
      todayXP: number;
      todayQuests: number;
      weeklyXP: number;
      level: number;
      isCurrentUser: boolean;
    }> = [];

    if (user.groupId) {
      // Get all group members
      const members = await ctx.db
        .query("users")
        .withIndex("by_group", (q) => q.eq("groupId", user.groupId))
        .collect();

      // Sort by weekly XP to find rank and leader
      const sorted = [...members].sort((a, b) => b.weeklyXp - a.weeklyXp);
      const userRank = sorted.findIndex((m) => m._id === userId) + 1;
      const leader = sorted[0];

      // Count active today (completed at least 1 quest today)
      const activeToday = new Set<string>();
      const todayGroupActivity = await ctx.db
        .query("activity")
        .withIndex("by_group_created", (q) => q.eq("groupId", user.groupId!))
        .filter((q) => q.gte(q.field("createdAt"), todayStart))
        .collect();

      for (const activity of todayGroupActivity) {
        if (activity.type === "quest_completed") {
          activeToday.add(activity.userId);
        }
      }

      // Calculate group's total XP today
      const groupTodayXP = todayGroupActivity
        .filter((a) => a.type === "quest_completed" && a.xp)
        .reduce((sum, a) => sum + (a.xp ?? 0), 0);

      groupStats = {
        memberCount: members.length,
        activeToday: activeToday.size,
        userRank,
        leaderName: leader?.name ?? "—",
        leaderXP: leader?.weeklyXp ?? 0,
        isUserLeading: leader?._id === userId,
        groupTodayXP,
      };

      // Get today's completed quests for each member
      for (const member of members) {
        const memberTodayQuests = await ctx.db
          .query("quests")
          .withIndex("by_user_created", (q) => q.eq("userId", member._id))
          .filter((q) => q.gte(q.field("createdAt"), todayStart))
          .collect();

        const memberTodayCompleted = memberTodayQuests.filter(
          (q) => q.status === "completed"
        );
        const memberTodayXP = memberTodayCompleted.reduce(
          (sum, q) => sum + q.xp,
          0
        );

        memberStats.push({
          name: member.name,
          todayXP: memberTodayXP,
          todayQuests: memberTodayCompleted.length,
          weeklyXP: member.weeklyXp,
          level: member.level,
          isCurrentUser: member._id === userId,
        });
      }
    }

    // Pick a random quote each time
    const quote = QUOTES[Math.floor(Math.random() * QUOTES.length)];

    return {
      today: {
        xp: todayXP,
        questsCompleted: todayCompleted.length,
        questsTotal: todayQuests.length,
      },
      week: {
        xp: user.weeklyXp,
        rank: groupStats?.userRank ?? 0,
      },
      group: groupStats,
      quote,
      memberStats,
      userName: user.name,
    };
  },
});

// Insight type for dynamic UI styling
type InsightType = "rivalry" | "analyst" | "stoic";

// Stats result type for action return
type StatsWithInsight = {
  today: { xp: number; questsCompleted: number; questsTotal: number };
  week: { xp: number; rank: number };
  group: {
    memberCount: number;
    activeToday: number;
    userRank: number;
    leaderName: string;
    leaderXP: number;
    isUserLeading: boolean;
    groupTodayXP: number;
  } | null;
  quote: string;
  memberStats: Array<{
    name: string;
    todayXP: number;
    todayQuests: number;
    weeklyXP: number;
    level: number;
    isCurrentUser: boolean;
  }>;
  userName: string;
  competitiveInsight: string;
  insightType: InsightType;
};

// Action to get dashboard with AI-generated competitive insight
export const getStatsWithInsight = action({
  args: { userId: v.id("users") },
  handler: async (ctx, { userId }): Promise<StatsWithInsight | null> => {
    // Get base stats from query
    const stats = await ctx.runQuery(api.dashboard.getStats, { userId });
    if (!stats) {
      return null;
    }

    // Generate competitive insight if we have group members
    let competitiveInsight = stats.quote;
    let insightType: InsightType = "stoic"; // Default to stoic mode

    if (stats.memberStats && stats.memberStats.length > 1) {
      try {
        const insight = await ctx.runAction(api.ai.generateGroupInsight, {
          members: stats.memberStats,
          currentUserName: stats.userName,
        });
        competitiveInsight = insight.insight;
        insightType = insight.type;
      } catch (e) {
        // Fall back to quote if AI fails (stoic mode)
        competitiveInsight = stats.quote;
        insightType = "stoic";
      }
    }

    return {
      ...stats,
      competitiveInsight,
      insightType,
    };
  },
});
