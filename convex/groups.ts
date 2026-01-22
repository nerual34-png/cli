import { v } from "convex/values";
import { mutation, query } from "./_generated/server";

// Create a new group
export const create = mutation({
  args: {
    name: v.string(),
    createdBy: v.id("users"),
  },
  handler: async (ctx, { name, createdBy }) => {
    const inviteCode = generateInviteCode();
    const now = Date.now();

    const groupId = await ctx.db.insert("groups", {
      name,
      inviteCode,
      createdBy,
      createdAt: now,
    });

    // Update user's group
    await ctx.db.patch(createdBy, {
      groupId,
      lastActiveAt: now,
    });

    return { groupId, inviteCode };
  },
});

// Get group by ID
export const get = query({
  args: { groupId: v.id("groups") },
  handler: async (ctx, { groupId }) => {
    return await ctx.db.get(groupId);
  },
});

// Get group by invite code
export const getByInviteCode = query({
  args: { inviteCode: v.string() },
  handler: async (ctx, { inviteCode }) => {
    return await ctx.db
      .query("groups")
      .withIndex("by_invite_code", (q) => q.eq("inviteCode", inviteCode.toUpperCase()))
      .unique();
  },
});

// Join a group by invite code
export const join = mutation({
  args: {
    userId: v.id("users"),
    inviteCode: v.string(),
  },
  handler: async (ctx, { userId, inviteCode }) => {
    const group = await ctx.db
      .query("groups")
      .withIndex("by_invite_code", (q) => q.eq("inviteCode", inviteCode.toUpperCase()))
      .unique();

    if (!group) {
      throw new Error("Invalid invite code");
    }

    const user = await ctx.db.get(userId);
    if (!user) {
      throw new Error("User not found");
    }

    if (user.groupId) {
      throw new Error("Already in a group");
    }

    const now = Date.now();

    // Update user's group
    await ctx.db.patch(userId, {
      groupId: group._id,
      lastActiveAt: now,
    });

    // Log activity
    await ctx.db.insert("activity", {
      groupId: group._id,
      userId,
      type: "joined_group",
      createdAt: now,
    });

    return { groupId: group._id, groupName: group.name };
  },
});

// Get group members
export const getMembers = query({
  args: { groupId: v.id("groups") },
  handler: async (ctx, { groupId }) => {
    return await ctx.db
      .query("users")
      .withIndex("by_group", (q) => q.eq("groupId", groupId))
      .collect();
  },
});

// Generate a unique invite code
function generateInviteCode(): string {
  const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789";
  let code = "";
  for (let i = 0; i < 6; i++) {
    code += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  return code.slice(0, 3) + "-" + code.slice(3);
}
