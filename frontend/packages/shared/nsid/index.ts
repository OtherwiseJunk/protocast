// Mirrors backend/internal/nsid. The prefix is a build variable, so nothing
// else in the frontend may spell an NSID literally.

export const PREFIX = "com.cacheblasters.protocast";

export function ref(name: string): string {
  return name === "" ? "" : `${PREFIX}.${name}`;
}

// Every schema file, including those that define no record.
export function schemaFiles(): string[] {
  return [ref("defs"), ref("show"), ref("episode")];
}

export function recordTypes(): string[] {
  return [ref("show"), ref("episode")];
}
