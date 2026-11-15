import { useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from './client'
import type { Label, ProjectDetail, Role, TaskQuery } from './types'

export const keys = {
  me: ['me'] as const,
  users: ['users'] as const,
  org: (id: number) => ['org', id] as const,
  orgProjects: (id: number) => ['org', id, 'projects'] as const,
  project: (id: number) => ['project', id] as const,
  projectDetail: (id: number) => ['project', id, 'detail'] as const,
  tasks: (id: number, query: TaskQuery) => ['project', id, 'tasks', query] as const,
  labels: (id: number) => ['project', id, 'labels'] as const,
  labelMap: (id: number, labelIds: number[]) => ['project', id, 'labelMap', labelIds] as const,
  stats: (id: number) => ['project', id, 'stats'] as const,
  workload: (id: number) => ['project', id, 'workload'] as const,
  allTasks: ['task'] as const,
  task: (id: number) => ['task', id] as const,
  activity: (id: number) => ['task', id, 'activity'] as const,
  tree: (id: number) => ['task', id, 'tree'] as const,
}

export function useMe() {
  return useQuery({ queryKey: keys.me, queryFn: api.me, staleTime: 60_000 })
}

export function useUsers() {
  return useQuery({ queryKey: keys.users, queryFn: api.listUsers })
}

export function useOrg(id: number | null) {
  return useQuery({ queryKey: keys.org(id ?? 0), queryFn: () => api.getOrg(id!), enabled: id !== null, retry: false })
}

export function useOrgProjects(id: number | null) {
  return useQuery({
    queryKey: keys.orgProjects(id ?? 0),
    queryFn: () => api.listOrgProjects(id!),
    enabled: id !== null,
    retry: false,
  })
}

export function useProject(id: number | null) {
  return useQuery({
    queryKey: keys.projectDetail(id ?? 0),
    queryFn: () => api.getProject(id!),
    enabled: id !== null,
    retry: false,
  })
}

export function useTasks(projectId: number, query: TaskQuery) {
  return useQuery({
    queryKey: keys.tasks(projectId, query),
    queryFn: () => api.listTasks(projectId, query),
    placeholderData: (prev) => prev,
  })
}

export function useLabels(projectId: number | null) {
  return useQuery({
    queryKey: keys.labels(projectId ?? 0),
    queryFn: () => api.listLabels(projectId!),
    enabled: projectId !== null,
  })
}

async function fetchTaskIdsForLabel(projectId: number, label: Label): Promise<number[]> {
  const ids: number[] = []
  let page = 1
  for (;;) {
    const res = await api.listTasks(projectId, { label: label.name, limit: 100, page })
    ids.push(...res.items.map((t) => t.id))
    if (page >= res.total_pages) break
    page += 1
  }
  return ids
}

export function useTaskLabelMap(projectId: number, labels: Label[] | undefined) {
  const labelIds = (labels ?? []).map((l) => l.id)
  return useQuery({
    queryKey: keys.labelMap(projectId, labelIds),
    enabled: labels !== undefined,
    placeholderData: (prev) => prev,
    queryFn: async () => {
      const map = new Map<number, Label[]>()
      const results = await Promise.all(
        (labels ?? []).map(async (label) => ({ label, ids: await fetchTaskIdsForLabel(projectId, label) })),
      )
      for (const { label, ids } of results) {
        for (const id of ids) {
          map.set(id, [...(map.get(id) ?? []), label])
        }
      }
      return map
    },
  })
}

export function useProjectStats(id: number) {
  return useQuery({ queryKey: keys.stats(id), queryFn: () => api.projectStats(id) })
}

export function useWorkload(id: number) {
  return useQuery({ queryKey: keys.workload(id), queryFn: () => api.projectWorkload(id) })
}

export function useTask(id: number) {
  return useQuery({ queryKey: keys.task(id), queryFn: () => api.getTask(id), retry: false })
}

export function useActivity(id: number) {
  return useQuery({ queryKey: keys.activity(id), queryFn: () => api.taskActivity(id) })
}

export function useTaskTree(id: number, enabled: boolean) {
  return useQuery({ queryKey: keys.tree(id), queryFn: () => api.taskTree(id), enabled })
}

export function useInvalidate() {
  const qc = useQueryClient()
  return {
    project: (id: number) => qc.invalidateQueries({ queryKey: keys.project(id) }),
    tasks: () => qc.invalidateQueries({ queryKey: keys.allTasks }),
    org: (id: number) => qc.invalidateQueries({ queryKey: keys.org(id) }),
    me: () => qc.invalidateQueries({ queryKey: keys.me }),
    users: () => qc.invalidateQueries({ queryKey: keys.users }),
    projectAndTasks: (id: number) =>
      Promise.all([
        qc.invalidateQueries({ queryKey: keys.project(id) }),
        qc.invalidateQueries({ queryKey: keys.allTasks }),
      ]),
  }
}

export function roleIn(project: ProjectDetail | undefined, userId: number | undefined): Role | null {
  if (!project || userId === undefined) return null
  return project.members.find((m) => m.user_id === userId)?.role ?? null
}

export interface Permissions {
  role: Role | null
  isMember: boolean
  canEdit: boolean
  isOwner: boolean
}

export function permissionsFor(project: ProjectDetail | undefined, userId: number | undefined): Permissions {
  const role = roleIn(project, userId)
  return {
    role,
    isMember: role !== null,
    canEdit: role === 'owner' || role === 'member',
    isOwner: role === 'owner',
  }
}
