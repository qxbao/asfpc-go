SELECT * FROM account WHERE "id" IN
(SELECT account_id from "group" WHERE group_id = '957808471691923')