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
	//"gopkg.in/DATA-DOG/go-sqlmock.v1"
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
		id                = "1"
		Url               = "abc.com"
		Crawl_timeout     = 10
		Frequency         = 20
		Failure_threshold = 3
		Status            = "active"
		Failure_count     = 0
	)

	s.mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT * FROM `url_infos` WHERE `url_infos`.`deleted_at` IS NULL AND ((id = ?))")).
		WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{"id", "url", "crawl_timeout", "frequency", "failure_threshold", "status", "failure_count"}).
			AddRow(id, Url, Crawl_timeout, Frequency, Failure_threshold, Status, Failure_count))

	res, err := s.interactor.Geturl(id)

	require.NoError(s.T(), err)
	require.Nil(s.T(), deep.Equal(Url, res.Url))
}

type AnyTime struct{}

func (s *Suite) TestPosturl() {
	var (
		id                = "1"
		Url               = "abc.com"
		Crawl_timeout     = 10
		Frequency         = 20
		Failure_threshold = 3
		Status            = "active"
		Failure_count     = 0
		//null              = "null"
	)
	record := UrlInfo{Url: "abc.com", Crawl_timeout: 10, Frequency: 20, Failure_threshold: 3, Status: "active", Failure_count: 0}
	record.ID = "1"
	s.mock.ExpectBegin()
	s.mock.ExpectExec(regexp.QuoteMeta(
		"INSERT INTO `url_infos`")).
		WithArgs(id, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), Url, Crawl_timeout, Frequency, Failure_threshold, Status, Failure_count).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()
	res, err := s.interactor.Inserturl(record)

	require.NoError(s.T(), err)
	require.Nil(s.T(), deep.Equal(Url, res.Url))

}

func (s *Suite) TestDeleteurl() {

	var (
		id = "1"
	)

	s.mock.ExpectBegin()
	s.mock.ExpectExec(regexp.QuoteMeta(
		"UPDATE `url_infos` SET `deleted_at`=? WHERE `url_infos`.`deleted_at` IS NULL AND `url_infos`.`id` = ?")).
		WithArgs(sqlmock.AnyArg(), id).
		WillReturnResult(sqlmock.NewResult(1, 1))

	s.mock.ExpectCommit()
	record := UrlInfo{Url: "abc.com", Crawl_timeout: 10, Frequency: 20, Failure_threshold: 3, Status: "active", Failure_count: 0}
	record.ID = "1"
	_, err := s.interactor.Deleteurl(record)

	require.NoError(s.T(), err)

}

func (s *Suite) TestActivateurl() {
	var (
		id     = "1"
		status = "active"
	)

	s.mock.ExpectBegin()
	s.mock.ExpectExec(regexp.QuoteMeta(
		"UPDATE `url_infos` SET `status` = ?, `updated_at` = ? WHERE `url_infos`.`deleted_at` IS NULL AND `url_infos`.`id` = ?")).
		WithArgs(status, sqlmock.AnyArg(), id).
		WillReturnResult(sqlmock.NewResult(1, 1))

	s.mock.ExpectCommit()
	record := UrlInfo{Url: "abc.com", Crawl_timeout: 10, Frequency: 20, Failure_threshold: 3, Status: "inactive", Failure_count: 0}
	record.ID = "1"
	err := s.interactor.Activateurl(record)

	require.NoError(s.T(), err)
}

func (s *Suite) TestDectivateurl() {
	var (
		id     = "1"
		status = "inactive"
	)

	s.mock.ExpectBegin()
	s.mock.ExpectExec(regexp.QuoteMeta(
		"UPDATE `url_infos` SET `status` = ?, `updated_at` = ? WHERE `url_infos`.`deleted_at` IS NULL AND `url_infos`.`id` = ?")).
		WithArgs(status, sqlmock.AnyArg(), id).
		WillReturnResult(sqlmock.NewResult(1, 1))

	s.mock.ExpectCommit()
	record := UrlInfo{Url: "abc.com", Crawl_timeout: 10, Frequency: 20, Failure_threshold: 3, Status: "active", Failure_count: 0}
	record.ID = "1"
	err := s.interactor.Deactivateurl(record)

	require.NoError(s.T(), err)
}

func (s *Suite) TestGetactiveurls() {
	var (
		id                = "1"
		status            = "active"
		Url               = "abc.com"
		Crawl_timeout     = 10
		Frequency         = 20
		Failure_threshold = 3
		Status            = "active"
		Failure_count     = 0
	)

	s.mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT * FROM `url_infos` WHERE `url_infos`.`deleted_at` IS NULL AND ((status = ?))")).
		WithArgs(status).
		WillReturnRows(sqlmock.NewRows([]string{"id", "url", "crawl_timeout", "frequency", "failure_threshold", "status", "failure_count"}).
			AddRow(id, Url, Crawl_timeout, Frequency, Failure_threshold, Status, Failure_count))

	urls, err := s.interactor.Getactiveurls()
	_ = urls

	require.NoError(s.T(), err)
}

func (s *Suite) TestUpdatecrawl() {

	var (
		id    = "1"
		crawl = 20
	)

	s.mock.ExpectBegin()
	s.mock.ExpectExec(regexp.QuoteMeta(
		"UPDATE `url_infos` SET `crawl_timeout` = ?, `updated_at` = ? WHERE `url_infos`.`deleted_at` IS NULL AND `url_infos`.`id` = ?")).
		WithArgs(crawl, sqlmock.AnyArg(), id).
		WillReturnResult(sqlmock.NewResult(1, 1))

	s.mock.ExpectCommit()
	record := UrlInfo{Url: "abc.com", Crawl_timeout: 10, Frequency: 20, Failure_threshold: 3, Status: "active", Failure_count: 0}
	record.ID = "1"
	err := s.interactor.Updatecrawl(record, crawl)

	require.NoError(s.T(), err)
}

func (s *Suite) TestUpdatefrequency() {

	var (
		id = "1"
		f  = 30
	)

	s.mock.ExpectBegin()
	s.mock.ExpectExec(regexp.QuoteMeta(
		"UPDATE `url_infos` SET `frequency` = ?, `updated_at` = ? WHERE `url_infos`.`deleted_at` IS NULL AND `url_infos`.`id` = ?")).
		WithArgs(f, sqlmock.AnyArg(), id).
		WillReturnResult(sqlmock.NewResult(1, 1))

	s.mock.ExpectCommit()
	record := UrlInfo{Url: "abc.com", Crawl_timeout: 10, Frequency: 20, Failure_threshold: 3, Status: "active", Failure_count: 0}
	record.ID = "1"
	err := s.interactor.Updatefrequency(record, f)

	require.NoError(s.T(), err)
}

func (s *Suite) TestUpdatethreshold() {

	var (
		id = "1"
		t  = 5
	)

	s.mock.ExpectBegin()
	s.mock.ExpectExec(regexp.QuoteMeta(
		"UPDATE `url_infos` SET `failure_threshold` = ?, `updated_at` = ? WHERE `url_infos`.`deleted_at` IS NULL AND `url_infos`.`id` = ?")).
		WithArgs(t, sqlmock.AnyArg(), id).
		WillReturnResult(sqlmock.NewResult(1, 1))

	s.mock.ExpectCommit()
	record := UrlInfo{Url: "abc.com", Crawl_timeout: 10, Frequency: 20, Failure_threshold: 3, Status: "active", Failure_count: 0}
	record.ID = "1"
	err := s.interactor.Updatethreshold(record, t)

	require.NoError(s.T(), err)
}

func (s *Suite) TestUpdatefailure() {

	var (
		id = "1"
		f  = 6
	)

	s.mock.ExpectBegin()
	s.mock.ExpectExec(regexp.QuoteMeta(
		"UPDATE `url_infos` SET `failure_count` = ?, `updated_at` = ? WHERE `url_infos`.`deleted_at` IS NULL AND `url_infos`.`id` = ?")).
		WithArgs(6, sqlmock.AnyArg(), id).
		WillReturnResult(sqlmock.NewResult(1, 1))

	s.mock.ExpectCommit()
	record := UrlInfo{Url: "abc.com", Crawl_timeout: 10, Frequency: 20, Failure_threshold: 3, Status: "active", Failure_count: 0}
	record.ID = "1"
	err := s.interactor.Updatefailure(record, f)

	require.NoError(s.T(), err)
}
