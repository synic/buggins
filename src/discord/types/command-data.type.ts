import { Interaction, SlashCommandBuilder } from 'discord.js';

export type CommandData = {
  data: SlashCommandBuilder;
  execute: (interaction: Interaction) => Promise<void>;
  requireMod?: boolean;
  autoreply?: boolean;
};
