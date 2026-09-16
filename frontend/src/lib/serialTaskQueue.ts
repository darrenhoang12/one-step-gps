// Keep preference writes in order so a slower request cannot overwrite a newer edit.
export function createSerialTaskQueue() {
  let tail: Promise<void> = Promise.resolve();

  return (task: () => Promise<void>): Promise<void> => {
    const result = tail.then(task);
    tail = result.then(
      () => {},
      () => {},
    );
    return result;
  };
}
