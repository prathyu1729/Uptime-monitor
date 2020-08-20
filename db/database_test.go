package db

import (
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-test/deep"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type Suite struct {
	suite.Suite
	DB   *gorm.DB
	mock sqlmock.Sqlmock

	interactor Dbinteraction
	url_infos  UrlInfo
}

func (s *Suite) SetupSuite() {
	var (
		db  *sql.DB
		err error
	)

	db, s.mock, err = sqlmock.New()
	require.NoError(s.T(), err)

	s.DB, err = gorm.Open("mysql", db)
	require.NoError(s.T(), err)

	s.DB.LogMode(true)

	s.interactor = CreateInteractor(s.DB)
}

func (s *Suite) AfterTest(_, _ string) {
	require.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func TestInit(t *testing.T) {
	suite.Run(t, new(Suite))
}

func (s *Suite) TestGeturl() {
	var (
		id                = int(1)
		Url               = "abc.com"
		Crawl_timeout     = 10
		Frequency         = 20
		Failure_threshold = 3
		Status            = "active"
		Failure_count     = 0
	)
	//s.SetupSuite()
	s.mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT * FROM `url_infos` WHERE  `url_infos`.`deleted_at` IS NULL AND ((id = ?))")).
		WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{"id", "url", "crawl_timeout", "frequency", "failure_threshold", "status", "failure_count"}).
			AddRow(id, Url, Crawl_timeout, Frequency, Failure_threshold, Status, Failure_count))

	res, err := s.interactor.Geturl(id)

	require.NoError(s.T(), err)
	require.Nil(s.T(), deep.Equal(Url, res.Url))
}

// func (s *Suite) TestPosturl() {
// 	var (
// 		id                = int(1)
// 		Url               = "abc.com"
// 		Crawl_timeout     = 10
// 		Frequency         = 20
// 		Failure_threshold = 3
// 		Status            = "active"
// 		Failure_count     = 0
// 	)
// 	//s.SetupSuite()
// 	s.mock.ExpectQuery(regexp.QuoteMeta(
// 		"SELECT * FROM `url_infos` WHERE  `url_infos`.`deleted_at` IS NULL AND ((id = ?))")).
// 		WithArgs(id).
// 		WillReturnRows(sqlmock.NewRows([]string{"id", "url", "crawl_timeout", "frequency", "failure_threshold", "status", "failure_count"}).
// 			AddRow(id, Url, Crawl_timeout, Frequency, Failure_threshold, Status, Failure_count))

// 	res, err := s.interactor.Geturl(id)

// 	require.NoError(s.T(), err)
// 	require.Nil(s.T(), deep.Equal(Url, res.Url))

// }

// func (s *Suite) TestActivateurl() {
// 	var (
// 		id                = int(1)
// 		Url               = "abc.com"
// 		Crawl_timeout     = 10
// 		Frequency         = 20
// 		Failure_threshold = 3
// 		Status            = "inactive"
// 		Failure_count     = 0
// 	)
// 	//s.SetupSuite()
// 	s.mock.MatchExpectationsInOrder(false)
// 	s.mock.ExpectBegin()
// 	s.mock.ExpectQuery(regexp.QuoteMeta(
// 		"SELECT * FROM `url_infos` WHERE  `url_infos`.`deleted_at` IS NULL AND ((id = ?))")).
// 		WithArgs(id).
// 		WillReturnRows(sqlmock.NewRows([]string{"id", "url", "crawl_timeout", "frequency", "failure_threshold", "status", "failure_count"}).
// 			AddRow(id, Url, Crawl_timeout, Frequency, Failure_threshold, Status, Failure_count))
// 	s.mock.ExpectCommit()

// 	s.mock.ExpectBegin()
// 	s.mock.ExpectQuery(regexp.QuoteMeta(
// 		"UPDATE `url_infos` set status=`active' WHERE ((id = ?)) ")).
// 		WithArgs(id)
// 	s.mock.ExpectCommit()

// 	s.mock.ExpectQuery(regexp.QuoteMeta(
// 		"UPDATE `url_infos` set failure_count=0 WHERE ((id = ?)) ")).
// 		WithArgs(id)
// 	s.mock.ExpectCommit()
// 	//WillReturnRows(sqlmock.NewRows([]string{"id", "url", "crawl_timeout", "frequency", "failure_threshold", "status", "failure_count"}).
// 	//AddRow(id, Url, Crawl_timeout, Frequency, Failure_threshold, Status, Failure_count))

// 	err := s.interactor.Activateurl(id)

// 	require.NoError(s.T(), err)

// }

// func (s *Suite) TestDeleteurl() {

// 	var (
// 		id                = int(1)
// 		Url               = "abc.com"
// 		Crawl_timeout     = 10
// 		Frequency         = 20
// 		Failure_threshold = 3
// 		Status            = "active"
// 		Failure_count     = 0
// 	)
// 	//s.SetupSuite()
// 	rows := sqlmock.NewRows([]string{"id", "url", "crawl_timeout", "frequency", "failure_threshold", "status", "failure_count"}).
// 		AddRow(id, Url, Crawl_timeout, Frequency, Failure_threshold, Status, Failure_count)

// 	s.mock.ExpectQuery(regexp.QuoteMeta(
// 		"SELECT * FROM `url_infos` WHERE  `url_infos`.`deleted_at` IS NULL AND ((id = ?))")).
// 		WithArgs(id).WillReturnRows(rows)
// 	//AddRow(id, Url, Crawl_timeout, Frequency, Failure_t

// 	s.mock.ExpectQuery(regexp.QuoteMeta(
// 		"DELETE FROM `url_infos` WHERE  `url_infos`.`deleted_at` IS NULL AND ((id = ?))")).
// 		WithArgs(id).WillReturnRows(rows)

// 	//WillReturnRows(sqlmock.NewRows([]string{"id", "url", "crawl_timeout", "frequency", "failure_threshold", "status", "failure_count"}).
// 	//AddRow(id, Url, Crawl_timeout, Frequency, Failure_threshold, Status, Failure_count))

// 	err := s.interactor.Deleteurl(id)

// 	require.NoError(s.T(), err)

// }
