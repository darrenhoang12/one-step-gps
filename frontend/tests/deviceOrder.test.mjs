import assert from "node:assert/strict";
import test from "node:test";
import { reorderVisibleDevices } from "../src/lib/deviceOrder.ts";

test("reordering visible cards preserves hidden slots and assigns unique positions", () => {
  const devices = [
    { device_id: "a", hidden: false, sort_order: 0 },
    { device_id: "hidden", hidden: true, sort_order: 1 },
    { device_id: "b", hidden: false, sort_order: 2 },
    { device_id: "c", hidden: false, sort_order: 3 },
  ];

  const reordered = reorderVisibleDevices(devices, ["c", "a", "b"]);
  assert.deepEqual(
    reordered.map(({ device_id, sort_order }) => [device_id, sort_order]),
    [["c", 0], ["hidden", 1], ["a", 2], ["b", 3]],
  );
  assert.deepEqual(
    reordered.map((device) => device.device_id),
    ["c", "hidden", "a", "b"],
  );
});

test("reordering rejects an incomplete or duplicate visible list", () => {
  const devices = [
    { device_id: "a", hidden: false, sort_order: 0 },
    { device_id: "hidden", hidden: true, sort_order: 1 },
    { device_id: "b", hidden: false, sort_order: 2 },
  ];

  assert.equal(reorderVisibleDevices(devices, ["b"]), null);
  assert.equal(reorderVisibleDevices(devices, ["b", "b"]), null);
  assert.equal(reorderVisibleDevices(devices, ["hidden", "a"]), null);
});
