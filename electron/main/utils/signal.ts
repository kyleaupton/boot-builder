import { deepMerge } from '@main/utils/merge';

type DeepPartial<T> = T extends object
  ? {
      [P in keyof T]?: DeepPartial<T[P]>;
    }
  : T;

export const signal = <T>({
  value,
  onUpdate,
}: {
  value: T;
  onUpdate?: (item: T) => void; // eslint-disable-line
}) => {
  let _value = value;

  return {
    get: () => {
      return _value;
    },
    set: (value: DeepPartial<T>) => {
      _value = deepMerge(_value, value);
      if (onUpdate) onUpdate(_value);
    },
  };
};

export type Signal<T> = ReturnType<typeof signal<T>>;
