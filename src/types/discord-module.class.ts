import { instance } from '@ao/storage';
import { Client, Events, Interaction, SlashCommandBuilder } from 'discord.js';
import { LocalStorage } from 'node-persist';
import { CommandData } from './command-data.type';

export abstract class DiscordModule {
  readonly _client: Client;
  readonly _commands = new Map<string, CommandData>();

  constructor(client: Client) {
    this._client = client;
    client.on(Events.InteractionCreate, async (interaction: Interaction) => {
      if (!interaction.isChatInputCommand()) return;

      const command = this._commands.get(interaction.commandName);

      try {
        await command?.execute(interaction);
        if (command?.autoreply) {
          await interaction.reply('Done!');
        }
      } catch (error) {
        console.error(error);
        if (interaction.replied || interaction.deferred) {
          await interaction.followUp({
            content: 'There was an error while executing this command!',
            ephemeral: true,
          });
        } else {
          await interaction.reply({
            content: 'There was an error while executing this command!',
            ephemeral: true,
          });
        }
      }
    });
  }

  get storage(): LocalStorage {
    return instance();
  }
  get client(): Client {
    return this._client;
  }
  get commands(): Map<string, CommandData> {
    return this._commands;
  }

  addCommand({
    name,
    description,
    execute,
    autoreply,
  }: {
    name: string;
    description: string;
    execute: (interaction: Interaction) => Promise<void>;
    autoreply?: boolean;
  }) {
    this._commands.set(name, {
      data: new SlashCommandBuilder().setName(name).setDescription(description),
      execute,
      autoreply,
    });
  }
}
