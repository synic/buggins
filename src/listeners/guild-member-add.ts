import { Client, GuildMember } from 'discord.js';

export default (client: Client): void => {
  client.on('guildMemberAdd', async (member: GuildMember) => {
    console.log(`${member.nickname} has joined`);

    const role = member.guild.roles.cache.find(
      (r) => r.name.toLowerCase() === 'photographers',
    );
    if (!role) {
      console.log('Could not find photographers role, not adding.');
      return;
    }

    try {
      await member.roles.add(role);
    } catch (error) {
      console.log(`error adding user to role: ${error}`);
    }
  });
};
