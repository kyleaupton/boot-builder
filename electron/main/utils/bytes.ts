export const humanReadableToBytes = (
  string: string,
): { bits: number; bytes: number } => {
  const units: Record<string, number> = {
    kiB: 1.024e3,
    MiB: 1.049e6,
    GiB: 1.07374e9,
    TiB: 1.09951e12,
    PiB: 1.1259e15,
  };

  const match = string.match(/(?<n>[\d\.]+) (?<unit>\w{3})/) // eslint-disable-line

  if (match?.groups) {
    const { n, unit } = match.groups;

    const factor = units[unit];

    if (factor) {
      const bytes = +n * factor;

      return {
        bits: bytes * 8,
        bytes,
      };
    }
  }

  return {
    bits: 0,
    bytes: 0,
  };
};
