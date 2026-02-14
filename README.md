# Heart beat monitoring

This project is go api which is part of uptimebeats.com heartbeat monitoring code


we have heartbeat schema mention in readme file at bottom

we have 2 main tasks

1. which will get pinged from url like below
https://heartbeat.uptimebeats.com/b/sdadsddaddqddwde
you need to find if that url exits in sites_list_heartbeat table if not then return one of then status code back saying id not found

if id is found then insert that details into table sites_status_heartbeat with correct details as per table's requirements

2. another go function which is running in background as cron functino every 1 min which

queries two table sites_status_heartbeat and sites_list_heartbeat and checks for every site 
if lastest record in sites_Status_heartbeats table for particular site is
sites_status_heartbeat.received_at time + sites_list_heartbeat.check_frequency+sites_list_heartbeat.toleranceDuration > current time then update the status(siteUp) of that site to down in sites_list_heartbeat table

and add record in incidents_heartbeat table for creating incident.


for above keep in mind to create single postgresql query with joins to it becomes faster then we can create incidents one by one or we can create function also if that's faster way. we just want faster and scalable way to do this. or you can suggest if i should run this as cron function directly insider postgresql using pg_cron extension in supabase. but keep not i want scalable and better way.








import { sql } from 'drizzle-orm';
import {
  bigint,
  boolean,
  index,
  integer,
  jsonb,
  pgTable,
  primaryKey,
  serial,
  smallint,
  text,
  timestamp,
  uuid,
} from 'drizzle-orm/pg-core';

export const organizationTable = pgTable(
  'organization',
  {
    id: uuid('id').notNull().defaultRandom().primaryKey(),
    name: text('name'),
    email: text('email').notNull(),
    image: text('image'),
    whatsappNumber: text('whatsapp_number'),
    telegramChatId: text('telegram_chat_id'),
    telegramChatUsername: text('telegram_chat_username'),
    restrictionId: bigint('restriction_id', { mode: 'number' })
      .references(() => plansRestrictionsTable.id)
      .default(1),
    emailList: text('email_list')
      .array()
      .notNull()
      .default(sql`'{}'::text[]`),
    discordWebhook: text('discord_webhook'),
    slackWebhook: text('slack_webhook'),
    teamsWebhook: text('teams_webhook'),
    webhookUri: text('webhook_uri'),
    timezone: text('timezone').notNull().default(''),
    pagerDutyKey: text('pagerduty_key'),
    googleChatWebhook: text('googlechat_webhook'),
    twilioSMSid: text('twilio_sms_id'),
    twilioPhoneNumber: text('twilio_phone_number'),
    twilioPhoneNumberSid: text('twilio_phone_number_sid'),
    twilioStatus: text('twilio_status').default('Not connected'),
    usersPhoneNumbers: text('users_phone_numbers')
      .array()
      .default(sql`'{}'::text[]`),
  },
  (table) => ({
    emailList: `CHECK (array_length(${table.emailList}, 1) < 3)`,
  })
);

export const siteUsersTable = pgTable('users', {
  id: uuid('id').notNull().primaryKey(),
  name: text('name'),
  email: text('email').notNull(),
  image: text('image'),
  timezone: text('timezone').notNull().default(''),
  currentOrg: uuid('current_org')
    .notNull()
    .references(() => organizationTable.id),
  restrictionId: bigint('restriction_id', { mode: 'number' })
    .references(() => plansRestrictionsTable.id)
    .default(1),
});

export const usersOrgRelTable = pgTable('users_org_rel', {
  id: bigint('id', { mode: 'number' }).generatedAlwaysAsIdentity().primaryKey(),
  createdAt: timestamp('created_at', { withTimezone: true })
    .notNull()
    .defaultNow(),
  userId: uuid('user_id')
    .notNull()
    .references(() => siteUsersTable.id, { onDelete: 'cascade' }),
  orgId: uuid('org_id')
    .notNull()
    .references(() => organizationTable.id, { onDelete: 'cascade' }),
  isAdmin: boolean('is_admin').notNull().default(false),
});

export const usersOrgRelInvitationsTable = pgTable(
  'users_org_rel_invitations',
  {
    id: bigint('id', { mode: 'number' })
      .generatedAlwaysAsIdentity()
      .primaryKey(),
    createdAt: timestamp('created_at', { withTimezone: true })
      .notNull()
      .defaultNow(),
    userId: uuid('user_id')
      .notNull()
      .references(() => siteUsersTable.id, { onDelete: 'cascade' }),
    orgId: uuid('org_id')
      .notNull()
      .references(() => organizationTable.id, { onDelete: 'cascade' }),
    isAdmin: boolean('is_admin').notNull().default(false),
    isAccepted: timestamp('is_accepted', { withTimezone: true }),
    isRejected: boolean('is_rejected'),
    invitorId: uuid('invitor_id').notNull(),
  }
);

export const plansTable = pgTable('plans', {
  id: serial('id').primaryKey(),
  productId: integer('productId').notNull(),
  productName: text('productName'),
  variantId: integer('variantId').notNull().unique(),
  name: text('name').notNull(),
  description: text('description'),
  price: text('price').notNull(),
  isUsageBased: boolean('isUsageBased').default(false),
  interval: text('interval'),
  intervalCount: integer('intervalCount'),
  trialInterval: text('trialInterval'),
  trialIntervalCount: integer('trialIntervalCount'),
  sort: integer('sort'),
});

export const webhookEventsTable = pgTable('webhookEvent', {
  id: integer('id').primaryKey(),
  createdAt: timestamp('createdAt', { mode: 'date' }).notNull().defaultNow(),
  eventName: text('eventName').notNull(),
  processed: boolean('processed').default(false),
  body: jsonb('body').notNull(),
  processingError: text('processingError'),
});

export const subscriptionsTable = pgTable('subscription', {
  id: serial('id').primaryKey(),
  lemonSqueezyId: text('lemonSqueezyId').unique().notNull(),
  orderId: integer('orderId').notNull(),
  name: text('name').notNull(),
  email: text('email').notNull(),
  status: text('status').notNull(),
  statusFormatted: text('statusFormatted').notNull(),
  renewsAt: text('renewsAt'),
  endsAt: text('endsAt'),
  trialEndsAt: text('trialEndsAt'),
  price: text('price').notNull(),
  isUsageBased: boolean('isUsageBased').default(false),
  isPaused: boolean('isPaused').default(false),
  subscriptionItemId: serial('subscriptionItemId'),
  userId: uuid('userId')
    .notNull()
    .references(() => siteUsersTable.id),
  planId: integer('planId')
    .notNull()
    .references(() => plansTable.id),
  orgId: uuid('org_id').references(() => organizationTable.id),
});

export const plansRestrictionsTable = pgTable('plans_restrictions', {
  id: bigint('id', { mode: 'number' })
    .generatedAlwaysAsIdentity()
    .notNull()
    .primaryKey(),
  createdAt: timestamp('created_at', { withTimezone: true })
    .notNull()
    .defaultNow(),
  httpMonitorCount: integer('http_monitor_count').default(10),
  httpMonitorFrequency: boolean('http_monitor_frequency')
    .notNull()
    .default(false),
  httpAdvance: boolean('http_advance').notNull().default(false),
  statusPageCount: integer('status_page_count').default(1),
  plansId: integer('plans_id').references(() => plansTable.id),
  name: text('name').notNull().unique(),
  dataRetension: integer('data_retension').notNull().default(30),
  loginSeatCount: smallint('login_seat_count').notNull().default(1),
});

// Schema for Status page table

export interface MonitorSectionItem {
  id: string;
  name: string;
  type: string;
  publicName: string;
  explanation?: string;
  widgetType: string;
  url: string;
  showUptime?: boolean;
  showResponseTime?: boolean;
  frequency?: number;
}

export interface MonitorSection {
  sectionName: string;
  sectionList: MonitorSectionItem[];
}

export const statusPageTable = pgTable('status_page', {
  id: uuid('id').notNull().default('gen_random_uuid()').primaryKey(),
  createdAt: timestamp('created_at', { withTimezone: true })
    .notNull()
    .defaultNow(),
  name: text('name').notNull(),
  statusPeriod: smallint('status_period').notNull().default(7),
  customDomain: text('custom_domain'),
  customDomainVerified: boolean('custom_domain_verified')
    .notNull()
    .default(false),
  logoLink: text('logo_link'),
  faviconLink: text('favicon_link'),
  ourLink: text('our_link'),
  pageStatus: boolean('page_status').notNull().default(true),
  showIncidents: boolean('show_incidents').notNull().default(false),
  showStatusUpdates: boolean('show_status_updates').notNull().default(false),
  logoLinkUrl: text('logo_link_url'),
  monitorSectionList: jsonb('monitor_section_list')
    .$type<MonitorSection>()
    .array(),
  getInTouch: text('get_in_touch'),
  currentTheme: boolean('current_theme').notNull().default(true), //light
  hideSearchEngine: boolean('hide_search_engine').notNull().default(false),
  googleAnalyticsId: text('google_analytics_id'),
  degradedValue: smallint('degraded_value').notNull().default(100),
  orgId: uuid('org_id').references(() => organizationTable.id),
});

// Status page updates table
export const statusPageUpdatesTable = pgTable('status_page_updates', {
  id: bigint('id', { mode: 'number' })
    .generatedAlwaysAsIdentity()
    .notNull()
    .primaryKey(),
  createdAt: timestamp('created_at', { withTimezone: true })
    .notNull()
    .defaultNow(),
  title: text('title').notNull(),
  message: text('message').notNull(),
  statusPageId: uuid('status_page_id')
    .notNull()
    .references(() => statusPageTable.id, {
      onDelete: 'cascade',
    }),
  monitorId: uuid('monitor_id').references(() => sitesTable.id, {
    onDelete: 'cascade',
  }),
  monitorStatus: text('monitor_status'),
});

// SSL and Domain expiry table
export const sslAndDomainExpiryTable = pgTable(
  'ssl_domain_expiry',
  {
    id: uuid('id').notNull().defaultRandom(),
    createdAt: timestamp('created_at', { withTimezone: true })
      .notNull()
      .defaultNow(),
    sslExpiry: timestamp('ssl_expiry', { withTimezone: true }),
    domainExpiry: timestamp('domain_expiry', { withTimezone: true }),
    url: text('url').notNull(),
    siteId: uuid('site_id')
      .notNull()
      .references(() => sitesTable.id, {
        onDelete: 'cascade',
        onUpdate: 'cascade',
      }),
  },
  (table) => ({
    pk: primaryKey({ columns: [table.id, table.url, table.siteId] }),
  })
);

// SSL and Domain expiry notifications logs table
export const sslAndDomainExpiryLogsTable = pgTable('ssl_domain_expiry_logs', {
  id: bigint('id', { mode: 'number' })
    .generatedAlwaysAsIdentity()
    .notNull()
    .primaryKey(),
  emailSent: timestamp('email_sent', { withTimezone: true }),
  sslDomainId: uuid('ssl_domain_id')
    .notNull()
    .references(() => sslAndDomainExpiryTable.id, {
      onDelete: 'cascade',
    }),
  type: boolean('type'),
});


// ////////////////////////////////////////////////////////////////////////////////////////
// HTTP 
// ////////////////////////////////////////////////////////////////////////////////////////
export const sitesTable = pgTable(
  'sites_list_http',
  {
    id: uuid('id').notNull().defaultRandom().primaryKey(),
    name: text('name'),
    type: text('type').notNull().default('http'),
    url: text('url').notNull(),
    createdAt: timestamp('created_at', { withTimezone: true })
      .notNull()
      .defaultNow(),
    updatedAt: timestamp('updated_at', { withTimezone: true })
      .notNull()
      .defaultNow(),
    checkFrequency: integer('check_frequency').notNull().default(300),
    responseTimeout: integer('response_timeout').notNull().default(30),
    requestHttpMethod: text('request_http_method').notNull().default('GET'),
    requestHeaders: jsonb('request_headers').notNull().default({}),
    responseHttpStatus: text('response_http_status')
      .notNull()
      .default('200-299'),
    requestBody: jsonb('request_body').default({}),
    siteUp: boolean('site_up').notNull().default(true),
    inMaintenance: boolean('in_maintenance').notNull().default(false),
    lastIncident: timestamp('last_incident', { withTimezone: true }),
    notifyTelegram: boolean('notify_telegram').notNull().default(false),
    notifyEmail: boolean('notify_email').notNull().default(false),
    notifyDiscord: boolean('notify_discord').notNull().default(false),
    notifySlack: boolean('notify_slack').notNull().default(false),
    notifyTeams: boolean('notify_teams').notNull().default(false),
    notifyWebhook: boolean('notify_webhook').notNull().default(false),
    notifyPagerDuty: boolean('notify_pagerduty').notNull().default(false),
    notifyGoogleChat: boolean('notify_googlechat').notNull().default(false),
    notifyTwilioSMS: boolean('notify_twilio_sms').notNull().default(false),
    pagerDutyDedupKey: text('pagerduty_dedup_key'),
    failureCount: integer('failure_count').notNull().default(0),
    locationList: text('location_list')
      .array()
      .notNull()
      .default(sql`'{weur}'::text[]`),
    orgId: uuid('org_id').references(() => organizationTable.id, {
      onDelete: 'cascade',
    }),
  },
  (table) => ({
    responseTimeoutCheck: `CHECK (${table.responseTimeout} <= 60)`,
  })
);

export const sitesStatusTable = pgTable(
  'sites_status_http',
  {
    id: bigint('id', { mode: 'number' })
      .generatedAlwaysAsIdentity()
      .notNull()
      .primaryKey(),
    createdAt: timestamp('created_at', { withTimezone: true })
      .notNull()
      .defaultNow(),
    siteName: text('site_name').notNull(),
    statusCode: bigint('status_code', { mode: 'number' }),
    responseTime: bigint('response_time', { mode: 'number' }),
    headers: text('headers'),
    siteUp: boolean('site_up').notNull().default(true),
    siteId: uuid('site_id').references(() => sitesTable.id, {
      onUpdate: 'cascade',
      onDelete: 'cascade',
    }),
    error: text('error'),
    location: text('location').default('weur'),
  },
  (table) => ({
    idxSitesStatusPerformance: index('idx_sites_status_performance').on(
      table.siteName,
      table.siteUp,
      table.createdAt
    ),
  })
);

export const incidentsSitesTable = pgTable('incidents_http', {
  id: bigint('id', { mode: 'number' })
    .generatedAlwaysAsIdentity()
    .notNull()
    .primaryKey(),
  createdAt: timestamp('created_at', { withTimezone: true })
    .notNull()
    .defaultNow(),
  status: boolean('status'),
  httpStatusId: bigint('http_status_id', { mode: 'number' })
    .notNull()
    .references(() => sitesStatusTable.id, {
      onDelete: 'cascade',
    }),
  httpSiteId: uuid('http_site_id')
    .notNull()
    .references(() => sitesTable.id, {
      onUpdate: 'cascade',
      onDelete: 'cascade',
    }),
  comments: text('comments').default(''),
  telegramSent: timestamp('telegram_notify_sent', { withTimezone: true }),
  emailSent: timestamp('email_notify_sent', { withTimezone: true }),
  discordSent: timestamp('discord_notify_sent', { withTimezone: true }),
  slackSent: timestamp('slack_notify_sent', { withTimezone: true }),
  teamsSent: timestamp('teams_notify_sent', { withTimezone: true }),
  webhookSent: timestamp('webhook_notify_sent', { withTimezone: true }),
  pagerDutySent: timestamp('pagerduty_notify_sent', { withTimezone: true }),
  googleChatSent: timestamp('googlechat_notify_sent', { withTimezone: true }),
  twilioSMSSent: timestamp('twilio_sms_notify_sent', { withTimezone: true }),
});

// ////////////////////////////////////////////////////////////////////////////////////////
// Ping 
// ////////////////////////////////////////////////////////////////////////////////////////
export const pingSitesTable = pgTable(
  'sites_list_ping',
  {
    id: uuid('id').notNull().defaultRandom().primaryKey(),
    name: text('name'),
    type: text('type').notNull().default('ping'),
    url_ip: text('url_ip').notNull(),
    createdAt: timestamp('created_at', { withTimezone: true })
      .notNull()
      .defaultNow(),
    updatedAt: timestamp('updated_at', { withTimezone: true })
      .notNull()
      .defaultNow(),
    checkFrequency: integer('check_frequency').notNull().default(300),
    responseTimeout: integer('response_timeout').notNull().default(30),
    siteUp: boolean('site_up').notNull().default(true),
    inMaintenance: boolean('in_maintenance').notNull().default(false),
    lastIncident: timestamp('last_incident', { withTimezone: true }),
    notifyTelegram: boolean('notify_telegram').notNull().default(false),
    notifyEmail: boolean('notify_email').notNull().default(false),
    notifyDiscord: boolean('notify_discord').notNull().default(false),
    notifySlack: boolean('notify_slack').notNull().default(false),
    notifyTeams: boolean('notify_teams').notNull().default(false),
    notifyWebhook: boolean('notify_webhook').notNull().default(false),
    notifyPagerDuty: boolean('notify_pagerduty').notNull().default(false),
    notifyGoogleChat: boolean('notify_googlechat').notNull().default(false),
    notifyTwilioSMS: boolean('notify_twilio_sms').notNull().default(false),
    pagerDutyDedupKey: text('pagerduty_dedup_key'),
    failureCount: integer('failure_count').notNull().default(0),
    orgId: uuid('org_id').references(() => organizationTable.id, {
      onDelete: 'cascade',
    }),
  },
  (table) => ({
    responseTimeoutCheck: `CHECK (${table.responseTimeout} <= 60)`,
  })
);

export const pingSitesStatusTable = pgTable(
  'sites_status_ping',
  {
    id: bigint('id', { mode: 'number' })
      .generatedAlwaysAsIdentity()
      .notNull()
      .primaryKey(),
    createdAt: timestamp('created_at', { withTimezone: true })
      .notNull()
      .defaultNow(),
    url_ip: text('url_ip').notNull(),
    responseTime: bigint('response_time', { mode: 'number' }),
    urlIpUp: boolean('url_ip_up').notNull().default(true),
    pingSiteId: uuid('ping_site_id').references(() => pingSitesTable.id, {
      onUpdate: 'cascade',
      onDelete: 'cascade',
    }),
    error: text('error'),
  },
  (table) => ({
    idxPingSitesStatusPerformance: index('idx_ping_sites_status_performance').on(
      table.url_ip,
      table.urlIpUp,
      table.createdAt
    ),
  })
);

export const incidentsPingSitesTable = pgTable('incidents_ping', {
  id: bigint('id', { mode: 'number' })
    .generatedAlwaysAsIdentity()
    .notNull()
    .primaryKey(),
  createdAt: timestamp('created_at', { withTimezone: true })
    .notNull()
    .defaultNow(),
  status: boolean('status'),
  pingStatusId: bigint('ping_status_id', { mode: 'number' })
    .notNull()
    .references(() => pingSitesStatusTable.id, {
      onDelete: 'cascade',
    }),
  pingSiteId: uuid('ping_site_id')
    .notNull()
    .references(() => pingSitesTable.id, {
      onUpdate: 'cascade',
      onDelete: 'cascade',
    }),
  comments: text('comments').default(''),
  telegramSent: timestamp('telegram_notify_sent', { withTimezone: true }),
  emailSent: timestamp('email_notify_sent', { withTimezone: true }),
  discordSent: timestamp('discord_notify_sent', { withTimezone: true }),
  slackSent: timestamp('slack_notify_sent', { withTimezone: true }),
  teamsSent: timestamp('teams_notify_sent', { withTimezone: true }),
  webhookSent: timestamp('webhook_notify_sent', { withTimezone: true }),
  pagerDutySent: timestamp('pagerduty_notify_sent', { withTimezone: true }),
  googleChatSent: timestamp('googlechat_notify_sent', { withTimezone: true }),
  twilioSMSSent: timestamp('twilio_sms_notify_sent', { withTimezone: true }),
});

// ////////////////////////////////////////////////////////////////////////////////////////
// Port 
// ////////////////////////////////////////////////////////////////////////////////////////
export const portSitesTable = pgTable(
  'sites_list_port',
  {
    id: uuid('id').notNull().defaultRandom().primaryKey(),
    name: text('name'),
    type: text('type').notNull().default('port'),
    url_ip: text('url_ip').notNull(),
    port: integer('port').notNull(),
    createdAt: timestamp('created_at', { withTimezone: true })
      .notNull()
      .defaultNow(),
    updatedAt: timestamp('updated_at', { withTimezone: true })
      .notNull()
      .defaultNow(),
    checkFrequency: integer('check_frequency').notNull().default(300),
    responseTimeout: integer('response_timeout').notNull().default(30),
    siteUp: boolean('site_up').notNull().default(true),
    inMaintenance: boolean('in_maintenance').notNull().default(false),
    lastIncident: timestamp('last_incident', { withTimezone: true }),
    notifyTelegram: boolean('notify_telegram').notNull().default(false),
    notifyEmail: boolean('notify_email').notNull().default(false),
    notifyDiscord: boolean('notify_discord').notNull().default(false),
    notifySlack: boolean('notify_slack').notNull().default(false),
    notifyTeams: boolean('notify_teams').notNull().default(false),
    notifyWebhook: boolean('notify_webhook').notNull().default(false),
    notifyPagerDuty: boolean('notify_pagerduty').notNull().default(false),
    notifyGoogleChat: boolean('notify_googlechat').notNull().default(false),
    notifyTwilioSMS: boolean('notify_twilio_sms').notNull().default(false),
    pagerDutyDedupKey: text('pagerduty_dedup_key'),
    failureCount: integer('failure_count').notNull().default(0),
    orgId: uuid('org_id').references(() => organizationTable.id, {
      onDelete: 'cascade',
    }),
  },
  (table) => ({
    responseTimeoutCheck: `CHECK (${table.responseTimeout} <= 60)`,
  })
);

export const portSitesStatusTable = pgTable(
  'sites_status_port',
  {
    id: bigint('id', { mode: 'number' })
      .generatedAlwaysAsIdentity()
      .notNull()
      .primaryKey(),
    createdAt: timestamp('created_at', { withTimezone: true })
      .notNull()
      .defaultNow(),
    url_ip: text('url_ip').notNull(),
    responseTime: bigint('response_time', { mode: 'number' }),
    urlIpUp: boolean('url_ip_up').notNull().default(true),
    portSiteId: uuid('port_site_id').references(() => portSitesTable.id, {
      onUpdate: 'cascade',
      onDelete: 'cascade',
    }),
    error: text('error'),
  },
  (table) => ({
    idxPortSitesStatusPerformance: index('idx_port_sites_status_performance').on(
      table.url_ip,
      table.urlIpUp,
      table.createdAt
    ),
  })
);

export const incidentsPortSitesTable = pgTable('incidents_port', {
  id: bigint('id', { mode: 'number' })
    .generatedAlwaysAsIdentity()
    .notNull()
    .primaryKey(),
  createdAt: timestamp('created_at', { withTimezone: true })
    .notNull()
    .defaultNow(),
  status: boolean('status'),
  portStatusId: bigint('port_status_id', { mode: 'number' })
    .notNull()
    .references(() => portSitesStatusTable.id, {
      onDelete: 'cascade',
    }),
  portSiteId: uuid('port_site_id')
    .notNull()
    .references(() => portSitesTable.id, {
      onUpdate: 'cascade',
      onDelete: 'cascade',
    }),
  comments: text('comments').default(''),
  telegramSent: timestamp('telegram_notify_sent', { withTimezone: true }),
  emailSent: timestamp('email_notify_sent', { withTimezone: true }),
  discordSent: timestamp('discord_notify_sent', { withTimezone: true }),
  slackSent: timestamp('slack_notify_sent', { withTimezone: true }),
  teamsSent: timestamp('teams_notify_sent', { withTimezone: true }),
  webhookSent: timestamp('webhook_notify_sent', { withTimezone: true }),
  pagerDutySent: timestamp('pagerduty_notify_sent', { withTimezone: true }),
  googleChatSent: timestamp('googlechat_notify_sent', { withTimezone: true }),
  twilioSMSSent: timestamp('twilio_sms_notify_sent', { withTimezone: true }),
});

// ////////////////////////////////////////////////////////////////////////////////////////
// DNS 
// ////////////////////////////////////////////////////////////////////////////////////////
export const dnsSitesTable = pgTable(
  'sites_list_dns',
  {
    id: uuid('id').notNull().defaultRandom().primaryKey(),
    name: text('name'),
    type: text('type').notNull().default('dns'),
    url: text('url').notNull(),
    createdAt: timestamp('created_at', { withTimezone: true })
      .notNull()
      .defaultNow(),
    updatedAt: timestamp('updated_at', { withTimezone: true })
      .notNull()
      .defaultNow(),
    checkFrequency: integer('check_frequency').notNull().default(300),
    expectedRecordList: jsonb('expected_record_list'),
    responseTimeout: integer('response_timeout').notNull().default(30),
    siteUp: boolean('site_up').notNull().default(true),
    inMaintenance: boolean('in_maintenance').notNull().default(false),
    lastIncident: timestamp('last_incident', { withTimezone: true }),
    notifyTelegram: boolean('notify_telegram').notNull().default(false),
    notifyEmail: boolean('notify_email').notNull().default(false),
    notifyDiscord: boolean('notify_discord').notNull().default(false),
    notifySlack: boolean('notify_slack').notNull().default(false),
    notifyTeams: boolean('notify_teams').notNull().default(false),
    notifyWebhook: boolean('notify_webhook').notNull().default(false),
    notifyPagerDuty: boolean('notify_pagerduty').notNull().default(false),
    notifyGoogleChat: boolean('notify_googlechat').notNull().default(false),
    notifyTwilioSMS: boolean('notify_twilio_sms').notNull().default(false),
    pagerDutyDedupKey: text('pagerduty_dedup_key'),
    failureCount: integer('failure_count').notNull().default(0),
    orgId: uuid('org_id').references(() => organizationTable.id, {
      onDelete: 'cascade',
    }),
  },
  (table) => ({
    responseTimeoutCheck: `CHECK (${table.responseTimeout} <= 60)`,
  })
);

export const dnsSitesStatusTable = pgTable(
  'sites_status_dns',
  {
    id: bigint('id', { mode: 'number' })
      .generatedAlwaysAsIdentity()
      .notNull()
      .primaryKey(),
    createdAt: timestamp('created_at', { withTimezone: true })
      .notNull()
      .defaultNow(),
    url: text('url').notNull(),
    urlUp: boolean('url_up').notNull().default(true),
    responseTime: bigint('response_time', { mode: 'number' }),
    actualRecordList: jsonb('actual_record_list'),
    dnsSiteId: uuid('dns_site_id').references(() => dnsSitesTable.id, {
      onUpdate: 'cascade',
      onDelete: 'cascade',
    }),
    error: text('error'),
  },
  (table) => ({
    idxDnsSitesStatusPerformance: index('idx_dns_sites_status_performance').on(
      table.url,
      table.createdAt
    ),
  })
);

export const incidentsDnsSitesTable = pgTable('incidents_dns', {
  id: bigint('id', { mode: 'number' })
    .generatedAlwaysAsIdentity()
    .notNull()
    .primaryKey(),
  createdAt: timestamp('created_at', { withTimezone: true })
    .notNull()
    .defaultNow(),
  status: boolean('status'),
  dnsStatusId: bigint('dns_status_id', { mode: 'number' })
    .notNull()
    .references(() => dnsSitesStatusTable.id, {
      onDelete: 'cascade',
    }),
  dnsSiteId: uuid('dns_site_id')
    .notNull()
    .references(() => dnsSitesTable.id, {
      onUpdate: 'cascade',
      onDelete: 'cascade',
    }),
  comments: text('comments').default(''),
  telegramSent: timestamp('telegram_notify_sent', { withTimezone: true }),
  emailSent: timestamp('email_notify_sent', { withTimezone: true }),
  discordSent: timestamp('discord_notify_sent', { withTimezone: true }),
  slackSent: timestamp('slack_notify_sent', { withTimezone: true }),
  teamsSent: timestamp('teams_notify_sent', { withTimezone: true }),
  webhookSent: timestamp('webhook_notify_sent', { withTimezone: true }),
  pagerDutySent: timestamp('pagerduty_notify_sent', { withTimezone: true }),
  googleChatSent: timestamp('googlechat_notify_sent', { withTimezone: true }),
  twilioSMSSent: timestamp('twilio_sms_notify_sent', { withTimezone: true }),
});


// ////////////////////////////////////////////////////////////////////////////////////////
// Heartbeat 
// ////////////////////////////////////////////////////////////////////////////////////////
export const heartbeatSitesTable = pgTable(
  'sites_list_heartbeat',
  {
    id: uuid('id').notNull().defaultRandom().primaryKey(),
    name: text('name'),
    type: text('type').notNull().default('heartbeat'),
    uniqueId: uuid('unique_id').notNull().defaultRandom(),
    createdAt: timestamp('created_at', { withTimezone: true })
      .notNull()
      .defaultNow(),
    updatedAt: timestamp('updated_at', { withTimezone: true })
      .notNull()
      .defaultNow(),
    checkFrequency: integer('check_frequency').notNull().default(300),
    toleranceDuration: integer('tolerance_duration').notNull().default(30),
    siteUp: boolean('site_up').notNull().default(true),
    inMaintenance: boolean('in_maintenance').notNull().default(false),
    lastIncident: timestamp('last_incident', { withTimezone: true }),
    notifyTelegram: boolean('notify_telegram').notNull().default(false),
    notifyEmail: boolean('notify_email').notNull().default(false),
    notifyDiscord: boolean('notify_discord').notNull().default(false),
    notifySlack: boolean('notify_slack').notNull().default(false),
    notifyTeams: boolean('notify_teams').notNull().default(false),
    notifyWebhook: boolean('notify_webhook').notNull().default(false),
    notifyPagerDuty: boolean('notify_pagerduty').notNull().default(false),
    notifyGoogleChat: boolean('notify_googlechat').notNull().default(false),
    notifyTwilioSMS: boolean('notify_twilio_sms').notNull().default(false),
    pagerDutyDedupKey: text('pagerduty_dedup_key'),
    failureCount: integer('failure_count').notNull().default(0),
    orgId: uuid('org_id').references(() => organizationTable.id, {
      onDelete: 'cascade',
    }),
  }
);

export const heartbeatSitesStatusTable = pgTable(
  'sites_status_heartbeat',
  {
    id: bigint('id', { mode: 'number' })
      .generatedAlwaysAsIdentity()
      .notNull()
      .primaryKey(),
    heartbeatSiteId: uuid('heartbeat_site_id').references(() => heartbeatSitesTable.id, {
      onUpdate: 'cascade',
      onDelete: 'cascade',
    }),
    received_at: timestamp('received_at', { withTimezone: true }),
    source_ip: text('source_ip'),
    http_method: text('http_method'),
    user_agent: text('user_agent'),
    status_code: bigint('status_code', { mode: 'number' }), // HTTP status code send by heartbeat
    error: text('error'),
  },
  (table) => ({
    idxHeartbeatSitesStatusPerformance: index('idx_heartbeat_sites_status_performance').on(
      table.received_at
    ),
  })
);

export const incidentsHeartbeatSitesTable = pgTable('incidents_heartbeat', {
  id: bigint('id', { mode: 'number' })
    .generatedAlwaysAsIdentity()
    .notNull()
    .primaryKey(),
  createdAt: timestamp('created_at', { withTimezone: true })
    .notNull()
    .defaultNow(),
  status: boolean('status'),
  heartbeatSiteId: uuid('heartbeat_site_id')
    .notNull()
    .references(() => heartbeatSitesTable.id, {
      onUpdate: 'cascade',
      onDelete: 'cascade',
    }),
  comments: text('comments').default(''),
  telegramSent: timestamp('telegram_notify_sent', { withTimezone: true }),
  emailSent: timestamp('email_notify_sent', { withTimezone: true }),
  discordSent: timestamp('discord_notify_sent', { withTimezone: true }),
  slackSent: timestamp('slack_notify_sent', { withTimezone: true }),
  teamsSent: timestamp('teams_notify_sent', { withTimezone: true }),
  webhookSent: timestamp('webhook_notify_sent', { withTimezone: true }),
  pagerDutySent: timestamp('pagerduty_notify_sent', { withTimezone: true }),
  googleChatSent: timestamp('googlechat_notify_sent', { withTimezone: true }),
  twilioSMSSent: timestamp('twilio_sms_notify_sent', { withTimezone: true }),
});



export type InsertStatusPageUpdate = typeof statusPageUpdatesTable.$inferInsert;
export type InsertStatusPage = typeof statusPageTable.$inferInsert;
export type InsertPlan = typeof plansTable.$inferInsert;
export type InsertWebhookEvent = typeof webhookEventsTable.$inferInsert;
export type InsertSubscription = typeof subscriptionsTable.$inferInsert;
export type InsertIncidentsSites = typeof incidentsSitesTable.$inferInsert;
export type InsertPlansRestrictions =
  typeof plansRestrictionsTable.$inferInsert;
export type InsertSslAndDomainExpiry =
  typeof sslAndDomainExpiryTable.$inferInsert;
export type InsertSslAndDomainExpiryLogs =
  typeof sslAndDomainExpiryLogsTable.$inferInsert;

