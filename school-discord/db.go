package schooldiscord

import (
	"github.com/Petrify/simp-core/service"
	simpsql "github.com/Petrify/simp-core/sql"
)

type modelModule struct {
	id    int
	name  string
	major string
	abbr  string
}

//a model final but minimized for searchability
type modelFinalSearchable struct {
	id     int
	name   string
	abbr   string
	majors []string
}

type modelFinal struct {
	id        int
	name      string
	abbr      string
	channelID string
	roleID    string
}

type modelUser struct {
	id       string
	finalIDs []int
}

func (s *Service) loadSettings(g *guild) error {
	tx, err := simpsql.UsingSchema(g.dbSchema)
	if err != nil {
		return err
	}

	row := tx.QueryRow("SELECT `value` FROM `option` WHERE `key` = 'command_prefix'")
	if err = row.Scan(&g.cmdPrefix); err != nil {
		return err
	}

	return nil
}

func (s *Service) getToken() error {

	tx, err := simpsql.UsingSchema(service.Schema(s))
	if err != nil {
		return err
	}

	row := tx.QueryRow("SELECT `value` FROM `option` WHERE `key` = 'token'")
	if err = row.Scan(&s.token); err != nil {
		return err
	}

	return nil
}

func (s *Service) getModuleCatalog(g *guild) ([]modelFinalSearchable, error) {

	tx, err := simpsql.UsingSchema(g.dbSchema)
	if err != nil {
		return nil, err
	}

	rows, err := tx.Query(
		`SELECT idfinal, module.name, module.abbr, module.fk_major FROM final
		JOIN module ON idfinal = module.fk_idfinal 
        GROUP BY idfinal, module.abbr, module.fk_major
        ORDER BY idfinal`)
	if err != nil {
		return nil, err
	}

	var lst []modelFinalSearchable = make([]modelFinalSearchable, 0)

	var tmpID int
	var lastModel = modelFinalSearchable{id: -1}
	var tmpName, tmpAbbr, tmpMajor string
	for rows.Next() {

		err := rows.Scan(&tmpID, &tmpName, &tmpAbbr, &tmpMajor)
		if err == nil {
			println(tmpID)
			if tmpID != lastModel.id {
				if lastModel.id != -1 {
					lst = append(lst, lastModel)
				}
				lastModel = modelFinalSearchable{
					id:     tmpID,
					name:   tmpName,
					abbr:   tmpAbbr,
					majors: make([]string, 0, 1),
				}
			}
			lastModel.majors = append(lastModel.majors, tmpMajor)
		}
	}

	return lst, nil
}

func (s *Service) getFinal(id int64, g *guild) (*modelFinal, error) {

	tx, err := simpsql.UsingSchema(g.dbSchema)
	if err != nil {
		return nil, err
	}

	rows, err := tx.Query(
		`SELECT idfinal, module.name, module.abbr, fk_idchannel, channel.fk_idrole FROM final
		JOIN module ON idfinal = module.fk_idfinal 
        LEFT JOIN channel ON fk_idchannel = channel.idchannel
        WHERE final.idfinal = ?
        GROUP BY idfinal`,
		id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	m := modelFinal{}

	if rows.Next() {
		var metaChanID, metaRoleID interface{}

		err = rows.Scan(&m.id, &m.name, &m.abbr, &metaChanID, &metaRoleID)
		if metaChanID == nil {
			m.channelID = ""
			m.roleID = ""
		} else {
			m.channelID = string(metaChanID.([]uint8))
			m.roleID = string(metaRoleID.([]uint8))
		}
		if err != nil {
			return nil, err
		}
		return &m, nil
	} else {
		return nil, nil
	}
}

func InsertRole(g *guild, roleID string) error {

	tx, err := simpsql.UsingSchema(g.dbSchema)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		`INSERT INTO role(idrole) VALUES (?)`, roleID)
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func insertChannel(g *guild, channelID string, roleID string) error {

	tx, err := simpsql.UsingSchema(g.dbSchema)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		`INSERT INTO channel(idchannel, fk_idrole) VALUES (?,?);`, channelID, roleID)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func setFinalChannel(g *guild, finalID int, channelID string) error {

	tx, err := simpsql.UsingSchema(g.dbSchema)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		`UPDATE final
		SET fk_idchannel = ?
		WHERE idfinal = ?`,
		channelID, finalID)
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func AddUserToFinal(g *guild, userID string, finalID int) error {

	tx, err := simpsql.UsingSchema(g.dbSchema)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		`INSERT INTO ref_user_has_final (iduser,idfinal)
		VALUES (?,?);`,
		userID, finalID)
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func newUser(g *guild, userID string) error {

	tx, err := simpsql.UsingSchema(g.dbSchema)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		`INSERT INTO user (iduser)
		VALUES (?);`,
		userID)
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func getUser(g *guild, userID string) (*modelUser, error) {

	tx, err := simpsql.UsingSchema(g.dbSchema)
	if err != nil {
		return nil, err
	}

	rows, err := tx.Query(
		`SELECT user.iduser, ref_user_has_final.idfinal FROM user
		LEFT JOIN ref_user_has_final ON ref_user_has_final.iduser = user.iduser
		WHERE user.iduser = ?`,
		userID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	defer rows.Close()

	var mu *modelUser

	for rows.Next() {
		mu = &modelUser{}
		mu.finalIDs = make([]int, 0)

		var metaFinalID interface{}
		rows.Scan(&mu.id, &metaFinalID)
		if metaFinalID != nil {
			mu.finalIDs = append(mu.finalIDs, int(metaFinalID.(int64)))
		}

	}

	tx.Commit()
	return mu, nil
}

func UserHasFinal(g *guild, userID string, finalID int) (bool, error) {

	tx, err := simpsql.UsingSchema(g.dbSchema)
	if err != nil {
		return false, err
	}

	rows, err := tx.Query(
		`SELECT iduser FROM ref_user_has_final
		WHERE iduser = ? AND idfinal = ?`,
		userID, finalID)
	if err != nil {
		tx.Rollback()
		return false, err
	}
	defer rows.Close()

	defer tx.Commit()
	return rows.Next(), nil
}

func RemoveUserFromFinal(g *guild, userID string, finalID int) error {

	tx, err := simpsql.UsingSchema(g.dbSchema)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		`DELETE FROM ref_user_has_final
		WHERE iduser = ? AND idfinal = ?;`,
		userID, finalID)
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func getUserFinals(g *guild, userID string) ([]modelFinal, error) {

	tx, err := simpsql.UsingSchema(g.dbSchema)
	if err != nil {
		return nil, err
	}

	rows, err := tx.Query(
		`SELECT ref.idfinal, module.name, module.abbr, final.fk_idchannel, channel.fk_idrole FROM ref_user_has_final AS ref
		LEFT JOIN final ON ref.idfinal = final.idfinal
		JOIN module ON ref.idfinal = module.fk_idfinal
		LEFT JOIN channel ON final.fk_idchannel = channel.idchannel
		WHERE ref.iduser = ?
		GROUP BY ref.idfinal;`,
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lst []modelFinal = make([]modelFinal, 0)

	for rows.Next() {
		mf := modelFinal{}
		var metaChanID, metaRoleID interface{}

		err = rows.Scan(&mf.id, &mf.name, &mf.abbr, &metaChanID, &metaRoleID)
		if metaChanID == nil {
			mf.channelID = ""
			mf.roleID = ""
		} else {
			mf.channelID = string(metaChanID.([]uint8))
			mf.roleID = string(metaRoleID.([]uint8))
		}
		if err == nil {
			lst = append(lst, mf)
		}
	}

	return lst, nil
}
