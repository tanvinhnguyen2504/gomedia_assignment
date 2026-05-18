TRUNCATE viewings RESTART IDENTITY;

INSERT INTO viewings (agent_id, lead_id, property_address, scheduled_at, status, notes) VALUES

-- SCHEDULED (future) — visible in normal list queries
(1, 10, '1 Orchard Blvd, #05-01, Singapore 238841',     NOW() + INTERVAL '2 days',   'SCHEDULED', 'Buyer wants morning slot'),
(1, 11, '88 Robinson Road, #12-00, Singapore 068912',    NOW() + INTERVAL '5 days',   'SCHEDULED', NULL),
(2, 12, '3 HarbourFront Walk, #08-22, Singapore 098499', NOW() + INTERVAL '1 day',    'SCHEDULED', 'Check carpark availability'),
(2, 13, '10 Bayfront Ave, #21-05, Singapore 018956',     NOW() + INTERVAL '10 days',  'SCHEDULED', NULL),
(3, 14, '5 Shenton Way, #30-01, Singapore 068808',       NOW() + INTERVAL '3 days',   'SCHEDULED', 'Bring floor plan'),

-- SCHEDULED but >1 hour old — MarkMissedViewings will flip this to MISSED on startup
(3, 10, '22 Sentosa Gateway, #04-10, Singapore 098135',  NOW() - INTERVAL '2 hours',  'SCHEDULED', 'No-show candidate'),
(3, 11, '33 Sentosa Gateway, #04-10, Vietnam 098135',  NOW() - INTERVAL '4 hours',  'SCHEDULED', 'No-show candidate'),

-- -- COMPLETED
(1, 12, '80 Pasir Panjang Rd, #03-11, Singapore 117372', NOW() - INTERVAL '7 days',  'COMPLETED', 'Client happy, making offer'),
(2, 10, '15 Queen St, #07-01, Singapore 188537',          NOW() - INTERVAL '14 days', 'COMPLETED', NULL),
(3, 13, '8 Kallang Ave, #02-05, Singapore 339509',        NOW() - INTERVAL '3 days',  'COMPLETED', 'Second viewing done'),

-- CANCELLED
(1, 13, '300 Beach Rd, #25-00, Singapore 199555',         NOW() - INTERVAL '5 days',  'CANCELLED', 'Client cancelled last minute'),
(2, 14, '1 Raffles Place, #40-01, Singapore 048616',      NOW() - INTERVAL '10 days', 'CANCELLED', NULL),

-- -- MISSED
(3, 11, '50 Collyer Quay, #18-02, Singapore 049321',      NOW() - INTERVAL '4 days',  'MISSED',    NULL),
(3, 1, '50 Test Quay, #18-02, Singapore 049321',      NOW() - INTERVAL '1 hour',  'MISSED',    NULL),
(1, 2, '111 Test Quay, #18-02, Singapore 049321',      NOW() - INTERVAL '2 hour',  'MISSED',    NULL)

