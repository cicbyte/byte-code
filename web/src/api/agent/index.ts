import { Alova } from '@/utils/http/alova/index';

/**
 * 外部 Agent 接入协议（dev-docs/agent-protocol.md）——前端仅用到接入码生成；
 * 注册/join/会话是 Agent 客户端（CLI/MCP）调用的，不出现在 Web 前端
 */

/** 生成项目接入码（owner/maintainer/超管；一次性/24h，只展示一次） */
export function createAgentJoinCode(projectId: number) {
  return Alova.Post<{ code: string; expiresAt: string }>(
    `/v1/projects/${projectId}/agent-codes`
  );
}

/** 调整 Agent 项目能力集（owner/maintainer；空数组=全部能力） */
export function updateAgentCapabilities(projectId: number, agentId: number, capabilities: string[]) {
  return Alova.Put<{}>(
    `/v1/projects/${projectId}/agents/${agentId}/capabilities`,
    { capabilities }
  );
}

/** Agent 能力字典（与后端 perm.AgentCaps 对应） */
export const AGENT_CAPS: Array<{ key: string; label: string }> = [
  { key: 'tasks_read', label: '读任务' },
  { key: 'tasks_write', label: '写任务' },
  { key: 'docs_read', label: '读文档' },
  { key: 'docs_write', label: '写文档' },
  { key: 'memory_read', label: '读记忆' },
  { key: 'memory_write', label: '写记忆' },
  { key: 'feedback', label: '投递反馈' },
  { key: 'qa', label: '维护问答库' },
];
