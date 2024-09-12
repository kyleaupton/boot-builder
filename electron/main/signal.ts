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
    set: (value: Partial<T>) => {
      _value = { ..._value, ...value };
      if (onUpdate) onUpdate(_value);
    },
  };
};
