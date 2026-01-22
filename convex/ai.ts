import { v } from "convex/values";
import { action } from "./_generated/server";
import { createVertex } from "@ai-sdk/google-vertex/edge";
import { generateText } from "ai";
import { api } from "./_generated/api";

// Group member stats for AI insight
interface MemberStats {
  name: string;
  todayXP: number;
  todayQuests: number;
  weeklyXP: number;
  level: number;
  isCurrentUser: boolean;
}

// Insight types for dynamic UI styling
type InsightType = "rivalry" | "analyst" | "stoic";

// Prompt templates for each insight mode
const RIVALRY_PROMPTS = [
  "{rival} is {gap} XP ahead. Do you really want to let them win?",
  "Gap to leader: {gap} XP. That's literally one quest. Do it.",
  "Second place is just the first loser. {quests} quest(s) to reclaim top spot.",
  "While you were resting, {rival} gained {rival_today} XP. Catch up.",
  "{rival} is only {gap} XP ahead. Target acquired.",
];

const ANALYST_PROMPTS = [
  "You're leading by {gap} XP. Don't let {rival} catch up.",
  "Optimal performance: {percent}% above your weekly average.",
  "Data shows {quests} quests secures your rank for 24 hours.",
  "Crew gained {group_xp} XP today. You contributed {percent}%.",
  "{rival} is closing in. {gap} XP gap shrinking.",
];

const STOIC_PROMPTS = [
  // Jungian / Shadow work
  "Where your fear is, there is your task.",
  "What you resist, persists. Face it.",
  "The cave you fear holds the treasure you seek.",
  "No tree can grow to heaven unless its roots reach down to hell.",

  // Stoic philosophy
  "The obstacle is the way.",
  "Waste no time arguing what a good man should be. Be one.",
  "You could leave life right now. Let that determine what you do.",
  "He who fears death will never do anything worthy of a living man.",

  // Nietzsche / Existential
  "He who has a why can bear almost any how.",
  "What doesn't kill me makes me stronger.",
  "You must have chaos within to give birth to a dancing star.",
  "Become who you are.",

  // Paradox / Depth
  "Whoever tries to save his life will lose it.",
  "The wound is where the light enters.",
  "Comfort is the enemy of progress.",
  "To live is to suffer. To survive is to find meaning.",

  // Action-oriented depth
  "Ship the fear. Debug later.",
  "Your ego is not your amigo.",
  "The grind reveals, it doesn't conceal.",
  "Do it scared. Do it anyway.",
];

// Determine insight mode based on user state
function determineInsightMode(
  members: MemberStats[],
  currentUserName: string
): InsightType {
  const sorted = [...members].sort((a, b) => b.weeklyXP - a.weeklyXP);
  const currentUser = members.find((m) => m.isCurrentUser);
  const userRank = sorted.findIndex((m) => m.isCurrentUser) + 1;
  const leader = sorted[0];

  if (!currentUser || members.length <= 1) {
    return "stoic";
  }

  // Rivalry: User is rank #2+ AND gap to leader is < 100 XP (catchable)
  if (
    userRank > 1 &&
    leader &&
    !leader.isCurrentUser &&
    leader.weeklyXP - currentUser.weeklyXP < 100
  ) {
    return "rivalry";
  }

  // Analyst: User is leading OR random 30% chance for data insight
  if (userRank === 1 || Math.random() < 0.3) {
    return "analyst";
  }

  // Default: Cyber-Stoic motivation
  return "stoic";
}

// Evaluate quest XP using Gemini
export const evaluateQuest = action({
  args: { title: v.string() },
  handler: async (ctx, { title }) => {
    const clientEmail = process.env.GOOGLE_CLIENT_EMAIL;
    const privateKey = process.env.GOOGLE_PRIVATE_KEY;
    const project = process.env.GOOGLE_CLOUD_PROJECT;

    if (!clientEmail) {
      throw new Error("Missing GOOGLE_CLIENT_EMAIL env var");
    }
    if (!privateKey) {
      throw new Error("Missing GOOGLE_PRIVATE_KEY env var");
    }
    if (!project) {
      throw new Error("Missing GOOGLE_CLOUD_PROJECT env var");
    }

    const vertex = createVertex({
      project,
      location: process.env.GOOGLE_CLOUD_LOCATION || "us-central1",
      googleCredentials: {
        clientEmail,
        privateKey: privateKey.replace(/\\n/g, "\n"),
      },
    });

    const { text } = await generateText({
      model: vertex("gemini-2.0-flash"),
      prompt: `You are an XP evaluator for a competitive productivity tracker. Users earn XP for ACTIVE effort that makes them better — coding, sports, learning, building, creating.

PHILOSOPHY:
This is a GRIND app. We reward EFFORT and ACTIVE work. Passive activities (sleep, rest, relaxation) get 0 XP — they're important for health but don't count as "grinding".

SCORING GUIDELINES:
- 0 XP: Passive/recovery (sleep, rest, nap, relax, chill) — NOT a grind task
- 5-15 XP: Trivial active tasks (reply to email, quick fix, short call)
- 20-40 XP: Small effort (reading 10-30 pages, routine workout, code review)
- 45-70 XP: Medium effort (deep work session, learning new skill, gym 1hr+)
- 75-100 XP: Large effort (ship feature, run 10km+, intense training)
- 100-150 XP: Epic (launch product, marathon, mass achievements)

WHAT COUNTS:
✓ Physical training (gym, running, sports)
✓ Learning (reading, courses, practice)
✓ Creating (coding, writing, building)
✓ Professional work (meetings, reviews, shipping)

WHAT DOESN'T COUNT (0 XP):
✗ Sleep, rest, nap, relaxation
✗ Passive consumption (watching TV, scrolling)
✗ Basic self-care (eating, showering)

Task: "${title}"

OUTPUT FORMAT (JSON only):
{
  "xp": <number 0-150>,
  "reasoning": "<brief explanation, 5-10 words>"
}`,
    });

    // Parse JSON response
    const jsonMatch = text.match(/\{[\s\S]*\}/);
    if (!jsonMatch) {
      throw new Error(`AI returned invalid response: ${text}`);
    }

    const result = JSON.parse(jsonMatch[0]);

    return {
      xp: Math.min(150, Math.max(0, result.xp)),
      reasoning: result.reasoning || "AI evaluated",
      needsClarification: false,
    };
  },
});

// Generate competitive insight about the group
export const generateGroupInsight = action({
  args: {
    members: v.array(
      v.object({
        name: v.string(),
        todayXP: v.number(),
        todayQuests: v.number(),
        weeklyXP: v.number(),
        level: v.number(),
        isCurrentUser: v.boolean(),
      })
    ),
    currentUserName: v.string(),
  },
  handler: async (
    ctx,
    { members, currentUserName }
  ): Promise<{ insight: string; type: InsightType; isAI: boolean }> => {
    // Determine insight mode based on user state
    const insightType = determineInsightMode(members, currentUserName);

    const clientEmail = process.env.GOOGLE_CLIENT_EMAIL;
    const privateKey = process.env.GOOGLE_PRIVATE_KEY;
    const project = process.env.GOOGLE_CLOUD_PROJECT;

    if (!clientEmail || !privateKey || !project) {
      // Fallback to template-based insight if AI not configured
      return generateFallbackInsight(members, currentUserName, insightType);
    }

    const vertex = createVertex({
      project,
      location: process.env.GOOGLE_CLOUD_LOCATION || "us-central1",
      googleCredentials: {
        clientEmail,
        privateKey: privateKey.replace(/\\n/g, "\n"),
      },
    });

    // Sort members by weekly XP for context
    const sorted = [...members].sort((a, b) => b.weeklyXP - a.weeklyXP);
    const currentUser = members.find((m) => m.isCurrentUser);
    const leader = sorted[0];

    const memberSummary = sorted
      .map(
        (m, i) =>
          `${i + 1}. ${m.name}${m.isCurrentUser ? " (you)" : ""}: ${m.weeklyXP} XP this week, ${m.todayXP} XP today, ${m.todayQuests} quests today, Level ${m.level}`
      )
      .join("\n");

    // Mode-specific prompt instructions
    const modeInstructions = {
      rivalry: `MODE: RIVALRY ALERT (Red alert - user is behind!)
Generate a message that makes them JEALOUS and creates URGENCY.
Examples:
- "alex is 40 XP ahead. Do you really want to let them win?"
- "Gap to leader: 25 XP. That's literally one quest. Do it."
- "Second place is just the first loser."`,

      analyst: `MODE: SYSTEM ANALYSIS (Data-driven insight)
Generate a message with SPECIFIC DATA about their performance.
Examples:
- "You're leading by 30 XP. Don't let mike catch up."
- "Crew gained 120 XP today. You contributed 40%."
- "Optimal performance: 20% above your weekly average."`,

      stoic: `MODE: GRIND MODE (Motivational/Cyberpunk)
Generate a SHORT, PUNCHY motivational message with hacker/cyberpunk vibes.
Examples:
- "Code is cheap. Show me the commit."
- "You didn't wake up today to be mediocre."
- "Focus. The noise is just input error."`,
    };

    const { text } = await generateText({
      model: vertex("gemini-2.0-flash"),
      prompt: `You are a competitive coach for a productivity group. Generate a SHORT, PUNCHY insight (max 60 chars).

${modeInstructions[insightType]}

GROUP LEADERBOARD (this week):
${memberSummary}

RULES:
- Match the MODE's tone and style
- Use the person's actual name when relevant
- Reference specific stats (XP gap, quests, etc)
- Keep it SHORT - max 60 characters!
- Make it feel aggressive and competitive

OUTPUT: Just the insight text, nothing else. Max 60 chars.`,
    });

    return {
      insight: text.trim().toLowerCase(),
      type: insightType,
      isAI: true,
    };
  },
});

// Fallback template-based insights when AI is not available
function generateFallbackInsight(
  members: MemberStats[],
  currentUserName: string,
  insightType: InsightType
): { insight: string; type: InsightType; isAI: boolean } {
  const sorted = [...members].sort((a, b) => b.weeklyXP - a.weeklyXP);
  const currentUser = members.find((m) => m.isCurrentUser);
  const leader = sorted[0];
  const currentUserRank = sorted.findIndex((m) => m.isCurrentUser) + 1;

  if (!currentUser || members.length <= 1) {
    // Solo user - use stoic mode
    const stoicInsight =
      STOIC_PROMPTS[Math.floor(Math.random() * STOIC_PROMPTS.length)];
    return { insight: stoicInsight, type: "stoic", isAI: false };
  }

  const gap = leader.isCurrentUser
    ? leader.weeklyXP - (sorted[1]?.weeklyXP || 0)
    : leader.weeklyXP - currentUser.weeklyXP;

  const rival = leader.isCurrentUser ? sorted[1] : leader;
  const questsNeeded = Math.ceil(gap / 40);
  const groupTodayXP = members.reduce((sum, m) => sum + m.todayXP, 0);
  const userPercent =
    groupTodayXP > 0
      ? Math.round((currentUser.todayXP / groupTodayXP) * 100)
      : 0;

  // Generate insight based on mode
  switch (insightType) {
    case "rivalry": {
      const template =
        RIVALRY_PROMPTS[Math.floor(Math.random() * RIVALRY_PROMPTS.length)];
      const insight = template
        .replace("{rival}", rival?.name || "the leader")
        .replace("{gap}", String(gap))
        .replace("{quests}", String(questsNeeded))
        .replace("{rival_today}", String(rival?.todayXP || 0));
      return { insight, type: "rivalry", isAI: false };
    }

    case "analyst": {
      const template =
        ANALYST_PROMPTS[Math.floor(Math.random() * ANALYST_PROMPTS.length)];
      const insight = template
        .replace("{rival}", rival?.name || "others")
        .replace("{gap}", String(gap))
        .replace("{quests}", String(questsNeeded))
        .replace("{group_xp}", String(groupTodayXP))
        .replace("{percent}", String(userPercent));
      return { insight, type: "analyst", isAI: false };
    }

    case "stoic":
    default: {
      const insight =
        STOIC_PROMPTS[Math.floor(Math.random() * STOIC_PROMPTS.length)];
      return { insight, type: "stoic", isAI: false };
    }
  }
}
