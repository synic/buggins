export const parseBool = (val?: string): boolean => {
  return ['t', 'true', '1'].includes(val?.toLowerCase() ?? '');
};
