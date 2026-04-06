export default () => ({
  port: parseInt(process.env.PORT ?? '3000', 10),
  collectorUrl: process.env.COLLECTOR_URL ?? 'http://localhost:8081',
  database: {
    url: process.env.POSTGRES_DSN ?? 'postgres://pulse:pulse_secret@localhost:5432/pulse_db',
  },
});
