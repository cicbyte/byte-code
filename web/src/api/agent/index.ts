import { Alova } from '@/utils/http/alova/index';

/**
 * 外部 Agent 接入协议（dev-docs/agent-protocol.md）——前端仅用到接入码生成；
 * 注册/join/会话是 Agent 客户端（CLI/MCP）调用的，不出现在 Web 前端
 */

/** 生成项目接入码（owner/超管；一次性/24h，只展示一次） */
export function createAgentJoinCode(projectId: number) {
  return Alova.Post<{ code: string; expiresAt: string }>(
    `/v1/projects/${projectId}/agent-codes`
  );
}
