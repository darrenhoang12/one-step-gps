import assert from "node:assert/strict";
import test from "node:test";
import { createSerialTaskQueue } from "../src/lib/serialTaskQueue.ts";

test("preference writes finish in the order they were queued", async () => {
  const enqueue = createSerialTaskQueue();
  const events = [];
  let finishFirst;
  const first = enqueue(async () => {
    events.push("first started");
    await new Promise((resolve) => { finishFirst = resolve; });
    events.push("first finished");
  });
  const second = enqueue(async () => {
    events.push("second started");
  });

  await Promise.resolve();
  assert.deepEqual(events, ["first started"]);
  finishFirst();
  await Promise.all([first, second]);
  assert.deepEqual(events, ["first started", "first finished", "second started"]);
});

test("a failed write does not block later writes", async () => {
  const enqueue = createSerialTaskQueue();
  const first = enqueue(async () => { throw new Error("save failed"); });
  let secondRan = false;
  const second = enqueue(async () => { secondRan = true; });

  await assert.rejects(first, /save failed/);
  await second;
  assert.equal(secondRan, true);
});
