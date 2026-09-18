import * as fsSync from 'fs';
import * as fs from 'fs/promises';

export async function saveData(path: string, data: string) {
    try {
        const fp = await fs.open(path, 'w+');
        await fp.writeFile(data);
        
        const content = await fs.readFile(path, 'utf-8');
        console.log(`${path} saved \n`, content);
        
        await fp.close();
        console.log('closed \n');
        return fp;
    } catch (err) {
        console.error(err);
    }
}

export function randomInt(): bigint {
    return BigInt(Math.floor(performance.timeOrigin * 1e6)) + process.hrtime.bigint();
}

export async function saveData2(path: string, data: string) {
    try {
        let tmp = `${path}.tmp.${randomInt()}`;
        let fp = await fs.open(tmp, 'w+');

        await fp.writeFile(data);
        
        const content = await fs.readFile(tmp, 'utf-8');
        console.log(`${tmp} saved \n`, content);
        
        await fp.close();
        console.log('closed \n');
        
        return fs.rename(tmp, path);
    } catch (err) {
        console.error(err);
    }
}

export function saveData3(path: string, data: string) {
    try {
        let tmp = `${path}.tmp.${randomInt()}`;
        

        let fd = fsSync.openSync(tmp, 'w+');

        fsSync.writeSync(fd, data);
        console.log(`${tmp} saved \n`, fsSync.readFileSync(tmp, 'utf-8'));
        
        fsSync.closeSync(fd);
        console.log('closed \n');
        
        return fsSync.renameSync(tmp, path);
    } catch (err) {
        console.error(err);
    }
}

export async function LogCreate(path: string) {
    return  await fs.open(path, 'a+');
}

export async function LogAppendAsync(fp: fs.FileHandle, line: string) {
    try {
        await fp.appendFile(line);
        await fp.sync();
        await fp.close();
    } catch (err) {
        console.error(err); 
    }
}

export function main() {
    const dir = 'data';
    
    if (!fsSync.existsSync(dir)) {
        fsSync.mkdirSync(dir);
    }
    
    const path = `${dir}/kvs`;
    const data = 'hello world \n';
    
    console.log('without atomic \n');
    
    saveData(path, data).then(() => {
        saveData2(path, data).then(() => {
            saveData3(path, data);
            });
        });
    LogCreate(path).then((fp) => {
        LogAppendAsync(fp, 'hello world \n');
        });
}
