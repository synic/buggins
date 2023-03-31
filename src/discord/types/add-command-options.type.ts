import { CommandData } from '.';

export type AddCommandOptions = Omit<CommandData, 'data'> & {
  name: string;
  description: string;
};
