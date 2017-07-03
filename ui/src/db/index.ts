import { Process, Project } from '../models';

function promisify<T>(req: IDBRequest): Promise<T> {
  return new Promise<T>((resolve, reject) => {
    req.onsuccess = (evt: any) => {
      resolve(req.result);
    };
    req.onerror = (evt: any) => {
      reject(req.error);
    };
  });
}

function promisifyCursor<T>(req: IDBRequest): Promise<Array<T>> {
  return new Promise<Array<T>>((resolve, reject) => {
    let arr = new Array<T>();
    req.onsuccess = (evt: any) => {
      if (req.result) {
        arr.push(req.result.value);
        req.result.continue();
      } else {
        resolve(arr);
      }
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

export const deleteProject = (id: number) => {
  return dbp.then(db =>
    promisify<undefined>(
      db.transaction('project', 'readwrite').objectStore('project').delete(id)
    )
  );
};
export const getProject = (id: number) => {
  return dbp.then(db =>
    promisify<Project>(
      db.transaction('project', 'readonly').objectStore('project').get(id)
    )
  );
};
export const getProjects = () => {
  return dbp.then(db =>
    promisifyCursor<Project>(
      db.transaction('project', 'readonly').objectStore('project').openCursor()
    )
  );
};
export const putProject = (project: Project) => {
  return dbp.then(db =>
    promisify<number>(
      db.transaction('project', 'readwrite').objectStore('project').put(project)
    )
  );
};

export const deleteProcess = (id: number) => {
  return dbp.then(db =>
    promisify<undefined>(
      db.transaction('process', 'readwrite').objectStore('process').delete(id)
    )
  );
};
export const getProcess = (processID: number) => {
  return dbp.then(db =>
    promisify<Process>(
      db
        .transaction('process', 'readonly')
        .objectStore('process')
        .get(processID)
    )
  );
};
export const getProcesses = (projectID: number) => {
  return dbp.then(db =>
    promisifyCursor<Process>(
      db
        .transaction('process', 'readonly')
        .objectStore('process')
        .index('projectID')
        .openCursor(projectID)
    )
  );
};
export const putProcess = (process: Process) => {
  return dbp.then(db =>
    promisify<number>(
      db.transaction('process', 'readwrite').objectStore('process').put(process)
    )
  );
};
