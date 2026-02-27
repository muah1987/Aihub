import { create } from 'zustand';
import { projectsApi, type Project } from '../api/projects';

interface ProjectState {
  projects: Project[];
  activeProject: Project | null;
  loading: boolean;
  error: string | null;
  fetchProjects: () => Promise<void>;
  setActiveProject: (project: Project | null) => void;
  createProject: (data: {
    name: string;
    description?: string;
    repo_owner: string;
    repo_name: string;
    provider_connection_id: string;
  }) => Promise<Project>;
  deleteProject: (id: string) => Promise<void>;
}

export const useProjectStore = create<ProjectState>((set, get) => ({
  projects: [],
  activeProject: null,
  loading: false,
  error: null,

  fetchProjects: async () => {
    set({ loading: true, error: null });
    try {
      const { data } = await projectsApi.list();
      set({ projects: data.projects || [], loading: false });
    } catch {
      set({ error: 'Failed to load projects', loading: false });
    }
  },

  setActiveProject: (project) => {
    set({ activeProject: project });
  },

  createProject: async (input) => {
    const { data } = await projectsApi.create(input);
    set({ projects: [...get().projects, data.project] });
    return data.project;
  },

  deleteProject: async (id) => {
    await projectsApi.delete(id);
    set({
      projects: get().projects.filter((p) => p.id !== id),
      activeProject: get().activeProject?.id === id ? null : get().activeProject,
    });
  },
}));
