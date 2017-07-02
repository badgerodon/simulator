import { Process, Project } from '../models';

interface eventedRequest<T> {
  result: T;
  error: DOMException;
  onsuccess: (evt: any) => void;
  onerror: (evt: any) => void;
}
function promisify<T>(req: eventedRequest<T>): Promise<T> {
  return new Promise<T>((resolve, reject) => {
    req.onsuccess = (evt: any) => {
      resolve(req.result);
    };
    req.onerror = (evt: any) => {
      reject(req.error);
    };
  });
}

let dbp = new Promise<IDBDatabase>((resolve, reject) => {
  const req = indexedDB.open('grpcsimulator', 1);
  req.onsuccess = (evt: any) => {
    resolve(req.result);
  };
  req.onerror = (evt: any) => {
    reject(evt);
  };
  req.onupgradeneeded = (evt: any) => {
    const projectStore = req.result.createObjectStore('project', {
      keyPath: 'id',
      autoIncrement: true
    });
    projectStore.createIndex('name', 'name', {
      unique: true
    });

    const processStore = req.result.createObjectStore('process', {
      keyPath: 'id',
      autoIncrement: true
    });
    processStore.createIndex('projectID', 'projectID', {
      unique: false
    });
    processStore.createIndex('name', ['projectID', 'config.name'], {
      unique: true
    });
  };
});

export const getProject = (id: number): Promise<Project> => {
  return dbp.then(db =>
    promisify(
      db.transaction('project', 'readonly').objectStore('project').get(id)
    )
  );
};
export const getProjects = (): Promise<Array<Project>> => {
  return dbp.then(db =>
    promisify(
      db.transaction('project', 'readonly').objectStore('project').getAll()
    )
  );
};
export const putProject = (project: Project): Promise<number> => {
  return dbp.then(db =>
    promisify(
      db.transaction('project', 'readwrite').objectStore('project').put(project)
    )
  );
};

export const putProcess = (process: Process): Promise<number> => {
  return dbp.then(db =>
    promisify(
      db.transaction('process', 'readwrite').objectStore('process').put(process)
    )
  );
};
export const getProcesses = (projectID: number): Promise<Array<Process>> => {
  return dbp.then(db =>
    promisify(
      db
        .transaction('process', 'readonly')
        .objectStore('process')
        .index('projectID')
        .getAll(projectID)
    )
  );
};

putProject({
  name: 'Hello World'
}).then(projectID =>
  putProcess({
    projectID: projectID,
    config: {
      name: 'echo-0',
      importPath: 'github.com/badgerodon/grpcsimulator/example/echo/server',
      branch: 'master'
    }
  })
);
