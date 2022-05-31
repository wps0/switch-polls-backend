CREATE TABLE IF NOT EXISTS `spolls_users` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    email VARCHAR(128) NOT NULL,
    create_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
INDEX ix_users_email(email)
);

CREATE TABLE IF NOT EXISTS `spolls_polls` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(256) NOT NULL,
    description VARCHAR(2048) NULL,
    create_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS `spolls_options` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    poll_id INT NOT NULL,
    content VARCHAR(1024) NOT NULL,
INDEX fk_options_poll_ix(poll_id),
FOREIGN KEY fk_options_poll_ix(poll_id)
    REFERENCES `spolls_polls`(id)
        ON DELETE CASCADE
        ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS `spolls_extras` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    option_id INT NOT NULL,
    type VARCHAR(64) NOT NULL,
    content VARCHAR(2048) NULL,
INDEX fk_extras_opt_ix (option_id),
FOREIGN KEY fk_extras_opt_ix(option_id)
    REFERENCES `spolls_options`(id)
        ON DELETE CASCADE
        ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS `spolls_votes` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    option_id INT NOT NULL,
    confirmed_at BIGINT NULL,
    create_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
INDEX fk_votes_usr_ix (user_id),
INDEX fk_votes_opt_ix (option_id),
FOREIGN KEY fk_votes_usr_ix(user_id)
    REFERENCES `spolls_users`(id)
        ON DELETE CASCADE
        ON UPDATE CASCADE,
FOREIGN KEY fk_votes_opt_ix(option_id)
    REFERENCES `spolls_options`(id)
        ON DELETE CASCADE
        ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS `spolls_confirmations` (
    token VARCHAR(192) NOT NULL PRIMARY KEY,
    vote_id INT NOT NULL,
    create_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
INDEX fk_confirmations_vote_id_ix (vote_id),
FOREIGN KEY fk_confirmations_vote_id_ix(vote_id)
    REFERENCES `spolls_votes`(id)
        ON DELETE CASCADE
        ON UPDATE CASCADE
);
