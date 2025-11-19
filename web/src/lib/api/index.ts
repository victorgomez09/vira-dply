import { apiClient as client } from './client';
import { authApi as auth } from './auth';
import { projectsApi as projects } from './projects';
import { environmentsApi as environments } from './environments';
import { applicationsApi as applications } from './applications';
import { databasesApi as databases } from './databases';
import { disksApi as disks } from './disks';
import { templatesApi as templates } from './templates';
import { studioApi as studio } from './studio';
import { activitiesApi as activities } from './activities';
import { serversApi as servers } from './servers';
import { organizationsApi as organizations } from './organizations';

export { client as apiClient };
export { auth as authApi };
export { projects as projectsApi };
export { environments as environmentsApi };
export { applications as applicationsApi };
export { databases as databasesApi };
export { disks as disksApi };
export { templates as templatesApi };
export { studio as studioApi };
export { activities as activitiesApi };
export { servers as serversApi };
export { organizations as organizationsApi };

export type { User, LoginRequest, RegisterRequest, AuthResponse } from './auth';
export type { Project, CreateProjectRequest } from './projects';
export type { Environment, CreateEnvironmentRequest } from './environments';
export type { Application, CreateApplicationRequest, ApplicationStatus } from './applications';
export type { Database, CreateDatabaseRequest, DatabaseType, DatabaseStatus } from './databases';
export type { Disk, CreateDiskRequest, ResizeDiskRequest, AttachDiskRequest } from './disks';
export type { ServiceTemplate, DeployTemplateRequest } from './templates';
export type { ApiError } from './client';
export type { Column, TableSchema, QueryResult, TableDataResult, Filter, Sort, DatabaseInfo, TableDataOptions, ExecuteQueryRequest, InsertRowRequest, UpdateRowRequest, DeleteRowRequest } from './studio';
export type { Activity, ActivitiesResponse } from './activities';
export type { Server, CreateServerRequest, UpdateServerRequest, ServersResponse } from './servers';
export type { Organization, OrganizationsResponse } from './organizations';

export const getProject = (id: string) => projects.get(id);
export const listProjects = () => projects.list();
export const createProject = (data: import('./projects').CreateProjectRequest) =>
	projects.create(data);
export const updateProject = (
	id: string,
	data: Partial<import('./projects').CreateProjectRequest>
) => projects.update(id, data);
export const deleteProject = (id: string) => projects.delete(id);

export const listEnvironments = (projectId: string) => environments.list(projectId);
export const getEnvironment = (projectId: string, id: string) => environments.get(projectId, id);
export const createEnvironment = (
	projectId: string,
	data: import('./environments').CreateEnvironmentRequest
) => environments.create(projectId, data);
export const updateEnvironment = (
	projectId: string,
	id: string,
	data: Partial<import('./environments').CreateEnvironmentRequest>
) => environments.update(projectId, id, data);
export const deleteEnvironment = (projectId: string, id: string) =>
	environments.delete(projectId, id);

export const listApplications = (projectId: string) => applications.list(projectId);
export const getApplication = (projectId: string, id: string) => applications.get(projectId, id);
export const createApplication = (
	projectId: string,
	data: import('./applications').CreateApplicationRequest
) => applications.create(projectId, data);
export const deleteApplication = (projectId: string, id: string) =>
	applications.delete(projectId, id);

export const listDatabases = (projectId: string, environmentId?: string) =>
	databases.list(projectId, environmentId);
export const getDatabase = (projectId: string, id: string) => databases.get(projectId, id);
export const createDatabase = (
	projectId: string,
	data: import('./databases').CreateDatabaseRequest
) => databases.create(projectId, data);
export const deleteDatabase = (projectId: string, id: string) => databases.delete(projectId, id);

export const listDisks = (projectId: string) => disks.list(projectId);
export const getDisk = (projectId: string, diskId: string) => disks.get(projectId, diskId);
export const createDisk = (projectId: string, data: import('./disks').CreateDiskRequest) =>
	disks.create(projectId, data);
export const resizeDisk = (projectId: string, diskId: string, data: import('./disks').ResizeDiskRequest) =>
	disks.resize(projectId, diskId, data);
export const deleteDisk = (projectId: string, diskId: string) => disks.delete(projectId, diskId);
export const attachDisk = (projectId: string, diskId: string, data: import('./disks').AttachDiskRequest) =>
	disks.attach(projectId, diskId, data);
export const detachDisk = (projectId: string, diskId: string) => disks.detach(projectId, diskId);

export const listTemplates = (category?: string) => templates.list(category);
export const getTemplate = (id: string) => templates.get(id);
export const deployTemplate = (templateId: string, data: import('./templates').DeployTemplateRequest) =>
	templates.deploy(templateId, data);
