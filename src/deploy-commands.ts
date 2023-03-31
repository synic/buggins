import { Client, REST, Routes } from 'discord.js';
import { ENABLED_MODULES } from './constants';

const rest = new REST({ version: '10' }).setToken(process.env.BOT_TOKEN ?? '');

const client = new Client({ intents: [] });

const commands = ENABLED_MODULES.map((moduleClass) => {
  const module = new moduleClass(client);
  return Array.from(module.commands.values());
}).flat(1);

(async () => {
  try {
    console.log(
      `Starting refeshing ${commands.length} application (/) commands...`,
    );
    const data = (await rest.put(
      Routes.applicationGuildCommands(
        process.env.CLIENT_ID ?? '',
        process.env.GUILD_ID ?? '',
      ),
      { body: commands.map((c) => c.data.toJSON()) },
    )) as { length: number };

    console.log(
      `Successfully reloaded ${data.length} application (/) commands.`,
    );
  } catch (error) {
    console.error(error);
    process.exit(1);
  }
})();
