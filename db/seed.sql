insert into augur_markets (id, category, title, description, address, author) values
(
  1,
  "binary",
  "Will an Asias team win The International DOTA2 Championships 2019?",
  'The International 2019 will be held in Shanghai at the Mercedes-Benz Arena.There will be 16 teams invited to the battle.The final match will happen in the 2019-07-03(fake, just for demo). If the champion is an asian team, then the market should resolve as "YES." Otherwise, the market should resolve as "NO." The definition of asian team is the country of the team belongs to Asian. Countries list is in accordance with <a href="https://en.wikipedia.org/wiki/List_of_Asian_countries_by_GDP." target="_blank">https://en.wikipedia.org/wiki/List_of_Asian_countries_by_GDP.</a>',
  "0x0000000000000000000000000000000000000001",
  "0x0000000000000000000000000000000000000001"
),
(
  2,
  "scalar",
  "What will be the size of The International DOTA2 Championships 2019 prize pool?",
  'Resolution will be based on data feed located at <a href="http://dota2.prizetrac.kr/" target="_blank">http://dota2.prizetrac.kr/</a>. The value can be found in the "Total Prize pool" column of the "Tournaments Stats" table, which will provide the total prize value at the moment before 12am UTC on July 20, 2019. ',
  "0x0000000000000000000000000000000000000002",
  "0x0000000000000000000000000000000000000002"
),
(
  3,
  "categorical",
  "Which team will win the 2019 DOTA2 TI competition?",
  'The market will resolve "Yes" for the winner of the 2019 DOTA2 TI competition set to start on 2019-07-03(fake, just for demo). The market will refer to <a href="https://www.dota2.com/international/overview/" target="_blank">https://www.dota2.com/international/overview/</a> for resolution.',
  "0x0000000000000000000000000000000000000003",
  "0x0000000000000000000000000000000000000003"
);

update augur_markets set minimum = "30000000", maximum = "80000000" where id = 2;

-- Tokens in ethereum-test-node

-- Augur Share Token address 0x179fd00c328d4ecdb5043c8686d377a24ede9d11
-- Augur Share Token address 0x114f5c774f8705f5b16b9fc494e3db6489f3237b
-- Augur Share Token address 0x1aa25040dbf401b3fdf67dcec5bb2fe2e531a55b
-- Augur Share Token address 0x36f179ff6e8a4816509ed867bd273fddeb409331
-- Augur Share Token address 0x9354c30a5d9f75785b711ddd3a7e134e1739b30a
-- Augur Share Token address 0xdda6e6b3ca7ed44ed3b8dc64047221ed994618fc
-- Augur Share Token address 0xf3c60116badca2c58e74e5fa8b20719284490c5e
-- Augur Share Token address 0x31e67d461d79835c271fd11aec73336a3a6dd6d7
-- Augur Share Token address 0xe05615f3b4cac6b1928b2a0ff31c0705e424a4bb

insert into tokens (address, symbol, name, decimals) values
("0x4c4fa7e8ea4cfcfc93deae2c0cff142a1dd3a218", "WETH",           "WETH",         18),

("0x179fd00c328d4ecdb5043c8686d377a24ede9d11", "1-LONG",         "LONG",         18),
("0x114f5c774f8705f5b16b9fc494e3db6489f3237b", "1-SHORT",        "SHORT",        18),

("0x1aa25040dbf401b3fdf67dcec5bb2fe2e531a55b", "2-LONG",         "LONG",         18),
("0x36f179ff6e8a4816509ed867bd273fddeb409331", "2-SHORT",        "SHORT",        18),

("0x9354c30a5d9f75785b711ddd3a7e134e1739b30a", "3-Virtus.Pro",   "Virtus.Pro",   18),
("0xdda6e6b3ca7ed44ed3b8dc64047221ed994618fc", "3-Secret",       "Secret",       18),
("0xf3c60116badca2c58e74e5fa8b20719284490c5e", "3-ViciGaming",   "ViciGaming",   18),
("0x31e67d461d79835c271fd11aec73336a3a6dd6d7", "3-EvilGeniuses", "EvilGeniuses", 18),
("0xe05615f3b4cac6b1928b2a0ff31c0705e424a4bb", "3-Other",        "Other",        18);

insert into markets (
  id,
  augur_market_id,
  base_token_address,
  base_token_decimals,
  base_token_symbol,
  base_token_name,
  quote_token_address,
  quote_token_decimals,
  quote_token_symbol,
  quote_token_name,
  min_order_size,
  maker_fee_rate,
  taker_fee_rate,
  price_precision,
  price_decimals,
  amount_decimals,
  gas_used_estimation,
  is_published,
  created_at
) values
("1-long",    1, "0x179fd00c328d4ecdb5043c8686d377a24ede9d11", 18, "1-LONG",         "LONG",         "0x4c4fa7e8ea4cfcfc93deae2c0cff142a1dd3a218", 18, "WETH", "WETH", 0.001, 0.003, 0.001, 5, 5, 5, 0, 1, datetime('now')),
("1-short",   1, "0x114f5c774f8705f5b16b9fc494e3db6489f3237b", 18, "1-SHORT",        "SHORT",        "0x4c4fa7e8ea4cfcfc93deae2c0cff142a1dd3a218", 18, "WETH", "WETH", 0.001, 0.003, 0.001, 5, 5, 5, 0, 1, datetime('now')),
("2-long",    2, "0x1aa25040dbf401b3fdf67dcec5bb2fe2e531a55b", 18, "2-LONG",         "LONG",         "0x4c4fa7e8ea4cfcfc93deae2c0cff142a1dd3a218", 18, "WETH", "WETH", 0.001, 0.003, 0.001, 5, 5, 5, 0, 1, datetime('now')),
("2-short",   2, "0x36f179ff6e8a4816509ed867bd273fddeb409331", 18, "2-SHORT",        "SHORT",        "0x4c4fa7e8ea4cfcfc93deae2c0cff142a1dd3a218", 18, "WETH", "WETH", 0.001, 0.003, 0.001, 5, 5, 5, 0, 1, datetime('now')),
("3-option1", 3, "0x9354c30a5d9f75785b711ddd3a7e134e1739b30a", 18, "3-Virtus.Pro",   "Virtus.Pro",   "0x4c4fa7e8ea4cfcfc93deae2c0cff142a1dd3a218", 18, "WETH", "WETH", 0.001, 0.003, 0.001, 5, 5, 5, 0, 1, datetime('now')),
("3-option2", 3, "0xdda6e6b3ca7ed44ed3b8dc64047221ed994618fc", 18, "3-Secret",       "Secret",       "0x4c4fa7e8ea4cfcfc93deae2c0cff142a1dd3a218", 18, "WETH", "WETH", 0.001, 0.003, 0.001, 5, 5, 5, 0, 1, datetime('now')),
("3-option3", 3, "0xf3c60116badca2c58e74e5fa8b20719284490c5e", 18, "3-ViciGaming",   "ViciGaming",   "0x4c4fa7e8ea4cfcfc93deae2c0cff142a1dd3a218", 18, "WETH", "WETH", 0.001, 0.003, 0.001, 5, 5, 5, 0, 1, datetime('now')),
("3-option4", 3, "0x31e67d461d79835c271fd11aec73336a3a6dd6d7", 18, "3-EvilGeniuses", "EvilGeniuses", "0x4c4fa7e8ea4cfcfc93deae2c0cff142a1dd3a218", 18, "WETH", "WETH", 0.001, 0.003, 0.001, 5, 5, 5, 0, 1, datetime('now')),
("3-option5", 3, "0xe05615f3b4cac6b1928b2a0ff31c0705e424a4bb", 18, "3-Other",        "Other",        "0x4c4fa7e8ea4cfcfc93deae2c0cff142a1dd3a218", 18, "WETH", "WETH", 0.001, 0.003, 0.001, 5, 5, 5, 0, 1, datetime('now'));