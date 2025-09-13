package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}


// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	// ожидаемое значение должно содержать Number из БД
	parcel.Number = id

	// get
	got, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, parcel, got)

	// delete
	require.NoError(t, store.Delete(id))
	_, err = store.Get(id)
	require.Error(t, err)
}


func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)
	parcel.Number = id

	// set address
	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)

	// check
	got, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newAddress, got.Address)

	// cleanup
	err = store.Delete(id)
	require.NoError(t, err)
}



func TestSetStatus(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err)
	parcel.Number = id

	newStatus := "sent"
	require.NoError(t, store.SetStatus(id, newStatus))

	// обновляем ожидаемое значение и сравниваем целиком
	parcel.Status = newStatus

	got, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, parcel, got)

	_ = store.Delete(id)
}


func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}

	// задаём всем посылкам одного клиента
	client := randRange.Intn(10_000_000)
	for i := range parcels {
		parcels[i].Client = client
	}

	// добавляем посылки и сохраняем в map по номеру
	parcelMap := make(map[int]Parcel)
	for i := range parcels {
		id, err := store.Add(parcels[i])
		require.NoError(t, err)
		require.NotZero(t, id)

		parcels[i].Number = id
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(storedParcels), len(parcels))

	// check
	for _, parcel := range storedParcels {
		expected, ok := parcelMap[parcel.Number]
		if !ok {
			continue // пропускаем чужие посылки этого же клиента, если такие есть
		}
		require.Equal(t, expected, parcel)
	}

	// cleanup
	for _, p := range parcels {
		err := store.Delete(p.Number)
		require.NoError(t, err)
	}
}
