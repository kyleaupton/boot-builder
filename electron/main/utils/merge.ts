/* eslint-disable @typescript-eslint/no-explicit-any */

export const deepMerge = <T, U>(target: T, source: U): T & U => {
  const output = { ...target };

  for (const key in source) {
    // @ts-ignore
    if (isObject(source[key]) && isObject(output[key])) {
      // If both source and target properties are objects, merge them recursively
      // @ts-ignore
      output[key] = deepMerge(output[key], source[key]);
    } else {
      // Otherwise, overwrite target's property with source's property
      // @ts-ignore
      output[key] = source[key];
    }
  }

  return output as T & U;
};

function isObject(obj: any): obj is object {
  return obj !== null && typeof obj === 'object';
}
