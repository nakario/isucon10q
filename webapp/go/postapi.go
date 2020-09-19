package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo"
)

// chair ////////////////////////////////////////////////////////////////////////////////

func postChair(c echo.Context) error {
	header, err := c.FormFile("chairs")
	if err != nil {
		c.Logger().Errorf("failed to get form file: %v", err)
		return c.NoContent(http.StatusBadRequest)
	}
	f, err := header.Open()
	if err != nil {
		c.Logger().Errorf("failed to open form file: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	defer f.Close()
	records, err := csv.NewReader(f).ReadAll()
	if err != nil {
		c.Logger().Errorf("failed to read csv: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	queryBuilder := &strings.Builder{}
	queryBuilder.WriteString("INSERT INTO chair(id, name, description, thumbnail, price, height, width, depth, color, features, kind, popularity, stock) VALUES ")
	queryParams := make([]interface{}, 0, len(records)*13)
	chairs := make([]Chair, 0)
	for _, row := range records {
		queryBuilder.WriteString("(?,?,?,?,?,?,?,?,?,?,?,?,?),")
		rm := RecordMapper{Record: row}
		id := rm.NextInt()
		name := rm.NextString()
		description := rm.NextString()
		thumbnail := rm.NextString()
		price := rm.NextInt()
		height := rm.NextInt()
		width := rm.NextInt()
		depth := rm.NextInt()
		color := rm.NextString()
		features := rm.NextString()
		kind := rm.NextString()
		popularity := rm.NextInt()
		stock := rm.NextInt()
		if err := rm.Err(); err != nil {
			c.Logger().Errorf("failed to read record: %v", err)
			return c.NoContent(http.StatusBadRequest)
		}
		chair := Chair{
			int64(id), name, description, thumbnail,
			int64(price), int64(height), int64(width), int64(depth),
			color, features, kind, int64(popularity), int64(stock),
		}
		chairs = append(chairs, chair)
		queryParams = append(queryParams, id, name, description, thumbnail, price, height, width, depth, color, features, kind, popularity, stock)
	}

	query := queryBuilder.String()[:queryBuilder.Len()-1] // remove trailing ","
	_, err = db_chair.Exec(query, queryParams...)
	if err != nil {
		c.Logger().Errorf("failed to insert chairs: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	for _, chair := range chairs {
		AddToChairSet(chair)
	}
	return c.NoContent(http.StatusCreated)
}

func buyChair(c echo.Context) error {
	m := echo.Map{}
	if err := c.Bind(&m); err != nil {
		c.Echo().Logger.Infof("post buy chair failed : %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	_, ok := m["email"].(string)
	if !ok {
		c.Echo().Logger.Info("post buy chair failed : email not found in request body")
		return c.NoContent(http.StatusBadRequest)
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Echo().Logger.Infof("post buy chair failed : %v", err)
		return c.NoContent(http.StatusBadRequest)
	}

	tx, err := db_chair.Beginx()
	if err != nil {
		c.Echo().Logger.Errorf("failed to create transaction : %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	defer tx.Rollback()

	var chair Chair
	err = tx.QueryRowx("SELECT * FROM chair WHERE id = ? AND stock > 0 FOR UPDATE", id).StructScan(&chair)
	if err != nil {
		if err == sql.ErrNoRows {
			c.Echo().Logger.Infof("buyChair chair id \"%v\" not found", id)
			return c.NoContent(http.StatusNotFound)
		}
		c.Echo().Logger.Errorf("DB Execution Error: on getting a chair by id : %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	_, err = tx.Exec("UPDATE chair SET stock = stock - 1 WHERE id = ?", id)
	if err != nil {
		c.Echo().Logger.Errorf("chair stock update failed : %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	err = tx.Commit()
	if err != nil {
		c.Echo().Logger.Errorf("transaction commit error : %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	AddToChairSet(chair)
	return c.NoContent(http.StatusOK)
}

// estate ////////////////////////////////////////////////////////////////////////////////

func postEstate(c echo.Context) error {
	header, err := c.FormFile("estates")
	if err != nil {
		c.Logger().Errorf("failed to get form file: %v", err)
		return c.NoContent(http.StatusBadRequest)
	}
	f, err := header.Open()
	if err != nil {
		c.Logger().Errorf("failed to open form file: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	defer f.Close()
	records, err := csv.NewReader(f).ReadAll()
	if err != nil {
		c.Logger().Errorf("failed to read csv: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	queryBuilder := &strings.Builder{}
	estates := make([]Estate, 0)
	queryBuilder.WriteString("INSERT INTO estate(id, name, description, thumbnail, address, latitude, longitude, geom, rent, door_height, door_width, features, popularity) VALUES ")
	queryParams := make([]interface{}, 0, len(records)*14)
	for _, row := range records {
		queryBuilder.WriteString("(?,?,?,?,?,?,?,ST_GeomFromText(CONCAT('POINT(',?,' ',?,')'), 4326),?,?,?,?,?),")

		rm := RecordMapper{Record: row}
		id := rm.NextInt()
		name := rm.NextString()
		description := rm.NextString()
		thumbnail := rm.NextString()
		address := rm.NextString()
		latitude := rm.NextFloat()
		longitude := rm.NextFloat()
		rent := rm.NextInt()
		doorHeight := rm.NextInt()
		doorWidth := rm.NextInt()
		features := rm.NextString()
		popularity := rm.NextInt()
		estate := Estate{
			int64(id), thumbnail,
			name, description,
			latitude, longitude,
			address, int64(rent),
			int64(doorHeight), int64(doorWidth),
			features, int64(popularity),
		}
		estates = append(estates, estate)
		if err := rm.Err(); err != nil {
			c.Logger().Errorf("failed to read record: %v", err)
			return c.NoContent(http.StatusBadRequest)
		}
		queryParams = append(queryParams, id, name, description, thumbnail, address, latitude, longitude, latitude, longitude, rent, doorHeight, doorWidth, features, popularity)
	}

	query := queryBuilder.String()[:queryBuilder.Len()-1] // remove trailing ","
	_, err = db_estate.Exec(query, queryParams...)
	if err != nil {
		c.Logger().Errorf("failed to insert estates: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	for _, estate := range estates {
		AddToEstateSet(estate)
	}
	return c.NoContent(http.StatusCreated)
}

func postEstateRequestDocument(c echo.Context) error {
	m := echo.Map{}
	if err := c.Bind(&m); err != nil {
		c.Echo().Logger.Infof("post request document failed : %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	_, ok := m["email"].(string)
	if !ok {
		c.Echo().Logger.Info("post request document failed : email not found in request body")
		return c.NoContent(http.StatusBadRequest)
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Echo().Logger.Infof("post request document failed : %v", err)
		return c.NoContent(http.StatusBadRequest)
	}
	got := GetEstate(int64(id))
	if got == nil {
		if err == sql.ErrNoRows {
			return c.NoContent(http.StatusNotFound)
		}
	}
	return c.NoContent(http.StatusOK)
}

func searchEstateNazotte(c echo.Context) error {
	coordinates := Coordinates{}
	err := c.Bind(&coordinates)
	if err != nil {
		c.Echo().Logger.Infof("post search estate nazotte failed : %v", err)
		return c.NoContent(http.StatusBadRequest)
	}

	if len(coordinates.Coordinates) == 0 {
		return c.NoContent(http.StatusBadRequest)
	}

	estatesInBoundingBox := []Estate{}
	query := fmt.Sprintf(`SELECT id, name, description, thumbnail, address, latitude, longitude, rent, door_height, door_width, features, popularity FROM estate WHERE ST_Contains(ST_PolygonFromText(%s, 4326), geom) ORDER BY popularity DESC, id ASC`, coordinates.coordinatesToText())
	err = db_estate.Select(&estatesInBoundingBox, query)
	if err == sql.ErrNoRows {
		c.Echo().Logger.Infof("select * from estate where ST_Contains ...", err)
		return NoIndentJSON(c, http.StatusOK, EstateSearchResponse{Count: 0, Estates: []Estate{}})
	} else if err != nil {
		c.Echo().Logger.Errorf("database execution error : %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	estatesInPolygon := estatesInBoundingBox

	var re EstateSearchResponse
	re.Estates = []Estate{}
	if len(estatesInPolygon) > NazotteLimit {
		re.Estates = estatesInPolygon[:NazotteLimit]
	} else {
		re.Estates = estatesInPolygon
	}
	re.Count = int64(len(re.Estates))

	return NoIndentJSON(c, http.StatusOK, re)
}
