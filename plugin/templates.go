package plugin

import (
	"google.golang.org/protobuf/compiler/protogen"
	"strings"
)

const placeholder = "Blarf"

func templateMessageType(source string, message *protogen.Message, trailingNewLine bool) string {
	result := strings.Replace(source, placeholder, getMessageProtoName(message), -1)
	if trailingNewLine {
		result += newLine
	}

	return result
}

var createTestTemplate = `
func TestCreateBlarf(t *testing.T) {
	fake := newFakeCreateBlarf()
	actual, err := createBlarf(fake)
	require.NoError(t, err)
	assertCreateBlarfEquality(t, fake, actual)
}`

var updateTestTemplate = `
func TestUpdateBlarf(t *testing.T) {
	fake := newFakeCreateBlarf()
	actual, err := createBlarf(fake)
	require.NoError(t, err)
	fakeUpdate := newFakeUpdateBlarf()
	updated, err := updateBlarf(actual.ID, fakeUpdate)
	require.NoError(t, err)
	require.Equal(t, actual.ID, updated.ID)
	assertUpdateBlarfEquality(t, fakeUpdate, updated)
	fetched, err := getBlarfById(actual.ID)
	require.NoError(t, err)
	assertBlarfByIdAfterUpdateEquality(t, fakeUpdate, fetched)
}`

var getByIdTestTemplate = `
func TestGetBlarfById(t *testing.T) {
	fake := newFakeCreateBlarf()
	actual, err := createBlarf(fake)
	require.NoError(t, err)
	fetched, err := getBlarfById(actual.ID)
	require.NoError(t, err)
	assertBlarfByIdAfterCreateEquality(t, actual, fetched)
}`

var deleteTestTemplate = `
func TestDeleteBlarf(t *testing.T) {
	fake := newFakeCreateBlarf()
	actual, err := createBlarf(fake)
	require.NoError(t, err)
	_, err = getBlarfById(actual.ID)
	require.NoError(t, err)
	err = deleteBlarf(actual.ID)
	require.NoError(t, err)
	_, err = getBlarfById(actual.ID)
	require.ErrorContains(t, err, "not found")
}`

var deleteFunctionTemplate = `
func deleteBlarf(id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := gqlClient.DeleteBlarf(ctx, id)
	return err
}`

var createFunctionTemplate = `
func createBlarf(createBlarf client.CreateBlarf_CreateBlarf) (client.CreateBlarf_CreateBlarf, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	response, err := gqlClient.CreateBlarf(ctx, %s)
	if err != nil {
		return client.CreateBlarf_CreateBlarf{}, err
	}
	return response.CreateBlarf, nil
}`

var updateFunctionTemplate = `
func updateBlarf(id uuid.UUID, updateBlarf client.UpdateBlarf_UpdateBlarf) (client.UpdateBlarf_UpdateBlarf, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	response, err := gqlClient.UpdateBlarf(ctx, id, %s)
	if err != nil {
		return client.UpdateBlarf_UpdateBlarf{}, err
	}
	return response.UpdateBlarf, nil
}`

var getByIdFunctionTemplate = `
func getBlarfById(id uuid.UUID) (*client.BlarfById_Blarf, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	response, err := gqlClient.BlarfByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return response.Blarf, nil
}`

var newFakeCreateFunctionTemplate = `
func newFakeCreateBlarf() client.CreateBlarf_CreateBlarf {
	fake := client.CreateBlarf_CreateBlarf{}
	%s
	return fake
}`

var newFakeUpdateFunctionTemplate = `
func newFakeUpdateBlarf() client.UpdateBlarf_UpdateBlarf {
	fake := client.UpdateBlarf_UpdateBlarf{}
	%s
	return fake
}`

var assertCreateEqualityTemplate = `
func assertCreateBlarfEquality(t *testing.T, expected, actual client.CreateBlarf_CreateBlarf) {
	%s
}`

var assertUpdateEqualityTemplate = `
func assertUpdateBlarfEquality(t *testing.T, expected, actual client.UpdateBlarf_UpdateBlarf) {
	%s
}`

var assertGetByIdAfterCreateEqualityTemplate = `
func assertBlarfByIdAfterCreateEquality(t *testing.T, expected client.CreateBlarf_CreateBlarf, actual *client.BlarfById_Blarf) {
	%s
}`

var assertGetByIdAfterUpdateEqualityTemplate = `
func assertBlarfByIdAfterUpdateEquality(t *testing.T, expected client.UpdateBlarf_UpdateBlarf, actual *client.BlarfById_Blarf) {
	%s
}`
