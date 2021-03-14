type MemoCheck = {
  name: string;
  timestamp: number;
};
const checkGroups = new Map<string, MemoCheck>();
const threshold = 30;
let checkEnabled = false;
setTimeout(() => {
  // checkEnabled = true;
}, 3000);
export const checkRenderingExclusive = (group: string, name: string) => {
  if (!checkEnabled) {
    return;
  }
  const now = Date.now();
  if (!checkGroups.has(group)) {
    checkGroups.set(group, { name, timestamp: now });
    return;
  }
  const latestCheck = checkGroups.get(group)!;
  if (latestCheck.name !== name && now - latestCheck.timestamp < threshold) {
    panic(`${latestCheck.name} and ${name} called in ${threshold}ms`);
  }
  latestCheck.name = name;
  latestCheck.timestamp = now;
};

export const panic = (message: string) => {
  document.getElementById("error")!.textContent = new Error(message).stack!;
};
