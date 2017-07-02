export interface WorkerConfiguration {
  name: string;
  importPath: string;
  branch: string;
}

export interface Project {
  id?: number;
  name: string;
}

export interface ProcessConfig {
  name: string;
  importPath: string;
  branch: string;
}
export interface Process {
  id?: number;
  projectID: number;
  config: ProcessConfig;
}
