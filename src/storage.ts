import { default as persist } from 'node-persist';

export async function init(): Promise<void> {
  await persist.init({
    dir: process.env.STORAGE_DIRECTORY ?? './data',
  });
}

export const instance = () => {
  if(!persist.defaultInstance) throw 'Could not initialize storage';
  return persist.defaultInstance;
};
