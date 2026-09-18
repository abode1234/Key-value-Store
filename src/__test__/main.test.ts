import { beforeAll, afterAll, describe, expect, it } from "vitest";
import { saveData, saveData2, saveData3 , LogAppendAsync, LogCreate } from "../index";
import * as fs from 'fs';

describe("Key-Value Store Persistence Tests", () => {
    beforeAll(() => {
        fs.mkdirSync('data', { recursive: true });
        fs.mkdirSync('data3', { recursive: true });
        fs.mkdirSync('data4', { recursive: true });
        fs.mkdirSync('data5', { recursive: true });

    });

    afterAll(() => {
        fs.rmSync('data', { recursive: true, force: true });
        fs.rmSync('data3', { recursive: true, force: true });
        fs.rmSync('data4', { recursive: true, force: true });
        fs.rmSync('data5', { recursive: true, force: true });
    });

    it("without atomic (saveData)", async () => {
        const data = 'hello world saveData\n';
        const path = 'data/kvs';

        await saveData(path, data);

        const content = fs.readFileSync(path, 'utf-8');
        expect(content).toBe(data);
    });

    it("with atomic async (saveData2)", async () => {
        const data = 'hello world saveData2\n';
        const path = 'data3/kvs';

        await saveData2(path, data);

        const content = fs.readFileSync(path, 'utf-8');
        expect(content).toBe(data);
    });

    it("with atomic sync (saveData3)", () => {
        const data = 'hello world saveData3 \n';
        const path = 'data4/kvs';

        saveData3(path, data);

        const content = fs.readFileSync(path, 'utf-8');
        expect(content).toBe(data);
    });

    it("with atomic sync (LogAppendAsync)", async () => {
        const data = 'hello world LogAppendAsync \n';
        const path = 'data5/kvs';
        const fp = await LogCreate(path);
        await LogAppendAsync(fp, data);
        await fp.close();

        const content = fs.readFileSync(path, 'utf-8');
        console.log(`${path} saved \n`, content);
        console.log('closed \n');
        expect(content).toBe(data);
    });
});
