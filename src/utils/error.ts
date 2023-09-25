export const getErrorMessage = (e: unknown) => {
  if (e instanceof Error) return e.message;
  else return String(e);
};
