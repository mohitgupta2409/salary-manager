// Package dto contains the API-layer Data Transfer Objects (DTOs) for the
// HTTP boundary of this service.
//
// DTOs are intentionally distinct from internal/model types:
//
//   - model.* structs describe how data is persisted in the database. They
//     carry foreign-key columns (e.g. CountryID, JobTitleID) and the raw
//     values stored on a row.
//   - dto.* structs describe what the HTTP API consumes and returns. They
//     present human-friendly, denormalised fields (e.g. country name and
//     currency instead of country_id) and shape input requests separately
//     from output responses.
//
// Each entity has its own file with three groups of types:
//
//  1. *CreateRequest / *UpdateRequest — what clients send in
//  2. *Response / *ListResponse       — what clients receive back
//  3. ToXxxResponse(...)              — converters from model -> dto
//     ToModelXxx(...)                 — converters from dto -> model
//
// The service layer is the boundary that performs the conversion: it
// accepts request DTOs from handlers, builds model values for the
// repository, and returns response DTOs back to the handlers.
package dto
