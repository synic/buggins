const dotenv = require('dotenv'); // jshint ignore:line
dotenv.config({ path: 'environment/local.env' });
dotenv.config({
  path: `environment/${process.env.NODE_ENV ?? 'development'}.env`,
});
