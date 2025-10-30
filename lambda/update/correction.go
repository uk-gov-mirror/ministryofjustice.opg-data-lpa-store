package main

import (
	"strconv"
	"time"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/validate"
	"github.com/ministryofjustice/opg-data-lpa-store/lambda/update/parse"
)

const signedAt = "/signedAt"

type Correction struct {
	Donor                   DonorPreRegistrationCorrection
	Attorney                AttorneyPreRegistrationCorrection
	CertificateProvider     CertificateProviderPreRegistrationCorrection
	AttorneyAppointmentType AttorneyAppointmentPreRegistrationCorrection
	TrustCorporation        TrustCorporationPreRegistrationCorrection
	AuthorisedSignatory     AuthorisedSignatoryPreRegistrationCorrection
	IndependentWitness      IndependentWitnessPreRegistrationCorrection
	WitnessedBy             WitnessedByPreRegistrationCorrection
	SignedAt                time.Time
}

type DonorPreRegistrationCorrection struct {
	shared.DonorCorrection
}

func (c DonorPreRegistrationCorrection) Apply(lpa *shared.Lpa) []shared.FieldError {
	isIdCheckComplete := lpa.Donor.IdentityCheck != nil
	isDobChangeRequested := !c.DateOfBirth.IsZero() && c.DateOfBirth != lpa.Donor.DateOfBirth

	if isIdCheckComplete && isDobChangeRequested {
		return []shared.FieldError{{
			Source: "/donor/dateOfBirth",
			Detail: "The donor's date of birth cannot be changed once the identity check is complete",
		}}
	}

	if lpa.Donor.FirstNames != c.FirstNames || lpa.Donor.LastName != c.LastName {
		donorNameChangeNote := shared.Note{
			Type:     "DONOR_NAME_CHANGE_V1",
			Datetime: time.Now().Format(time.RFC3339),
			Values: map[string]string{
				"newName": c.FirstNames + " " + c.LastName,
			},
		}

		lpa.AddNote(donorNameChangeNote)
	}

	if lpa.Donor.DateOfBirth != c.DateOfBirth {
		donorDobChangeNote := shared.Note{
			Type:     "DONOR_DOB_CHANGE_V1",
			Datetime: time.Now().Format(time.RFC3339),
			Values: map[string]string{
				"newDob": c.DateOfBirth.DateOnlyText(),
			},
		}

		lpa.AddNote(donorDobChangeNote)
	}

	lpa.Donor.FirstNames = c.FirstNames
	lpa.Donor.LastName = c.LastName
	lpa.Donor.OtherNamesKnownBy = c.OtherNamesKnownBy
	lpa.Donor.DateOfBirth = c.DateOfBirth
	lpa.Donor.Address = c.Address
	lpa.Donor.Email = c.Email

	return nil
}

type CertificateProviderPreRegistrationCorrection struct {
	shared.CertificateProviderCorrection
}

func (c CertificateProviderPreRegistrationCorrection) Apply(lpa *shared.Lpa) []shared.FieldError {
	if !c.SignedAt.IsZero() && !c.SignedAt.Equal(*lpa.CertificateProvider.SignedAt) && lpa.Channel == shared.ChannelOnline {
		return []shared.FieldError{{
			Source: "/certificateProvider" + signedAt,
			Detail: "The Certificate Provider Signed on date cannot be changed for online LPAs",
		}}
	}

	if lpa.CertificateProvider.FirstNames != c.FirstNames || lpa.CertificateProvider.LastName != c.LastName {
		nameChangeNote := shared.Note{
			Type:     "CERTIFICATE_PROVIDER_NAME_CHANGE_V1",
			Datetime: time.Now().Format(time.RFC3339),
			Values: map[string]string{
				"newName": c.FirstNames + " " + c.LastName,
			},
		}

		lpa.AddNote(nameChangeNote)
	}

	lpa.CertificateProvider.FirstNames = c.FirstNames
	lpa.CertificateProvider.LastName = c.LastName
	lpa.CertificateProvider.Address = c.Address
	lpa.CertificateProvider.Email = c.Email
	lpa.CertificateProvider.Phone = c.Phone

	if !c.SignedAt.IsZero() {
		lpa.CertificateProvider.SignedAt = &c.SignedAt
	}

	return nil
}

type AttorneyPreRegistrationCorrection struct {
	shared.AttorneyCorrection
}

func (c AttorneyPreRegistrationCorrection) Apply(lpa *shared.Lpa) []shared.FieldError {
	if c.Index != nil {
		if !c.SignedAt.IsZero() && !c.SignedAt.Equal(*lpa.Attorneys[*c.Index].SignedAt) && lpa.Channel == shared.ChannelOnline {
			source := "/attorney/" + strconv.Itoa(*c.Index) + signedAt
			return []shared.FieldError{{Source: source, Detail: "The attorney signed at date cannot be changed for online LPA"}}
		}

		attorney := &lpa.Attorneys[*c.Index]
		attorney.FirstNames = c.FirstNames
		attorney.LastName = c.LastName
		attorney.DateOfBirth = c.DateOfBirth
		attorney.Address = c.Address
		attorney.Email = c.Email
		attorney.Mobile = c.Mobile
		attorney.SignedAt = &c.SignedAt
	}

	return nil
}

type TrustCorporationPreRegistrationCorrection struct {
	shared.TrustCorporationCorrection
}

func (c TrustCorporationPreRegistrationCorrection) Apply(lpa *shared.Lpa) []shared.FieldError {
	if c.Index != nil {
		trustCorporation := &lpa.TrustCorporations[*c.Index]

		trustCorporation.Name = c.Name
		trustCorporation.CompanyNumber = c.CompanyNumber
		trustCorporation.Email = c.Email
		trustCorporation.Address = c.Address
		trustCorporation.Mobile = c.Mobile

		for i, tcs := range c.Signatories {
			trustCorporation.Signatories[i].FirstNames = tcs.FirstNames
			trustCorporation.Signatories[i].LastName = tcs.LastName
			trustCorporation.Signatories[i].ProfessionalTitle = tcs.ProfessionalTitle
		}
	}

	return nil
}

type AttorneyAppointmentPreRegistrationCorrection struct {
	shared.AttorneyAppointmentTypeCorrection
}

func (c AttorneyAppointmentPreRegistrationCorrection) Apply(lpa *shared.Lpa) []shared.FieldError {
	if !c.HowAttorneysMakeDecisions.Unset() {
		lpa.HowAttorneysMakeDecisions = c.HowAttorneysMakeDecisions
		lpa.HowAttorneysMakeDecisionsIsDefault = false
	}

	if !c.HowReplacementAttorneysMakeDecisions.Unset() {
		lpa.HowReplacementAttorneysMakeDecisions = c.HowReplacementAttorneysMakeDecisions
		lpa.HowReplacementAttorneysMakeDecisionsIsDefault = false
	}

	if !c.LifeSustainingTreatmentOption.Unset() {
		lpa.LifeSustainingTreatmentOption = c.LifeSustainingTreatmentOption
		lpa.LifeSustainingTreatmentOptionIsDefault = false
	}

	if !c.WhenTheLpaCanBeUsed.Unset() {
		lpa.WhenTheLpaCanBeUsed = c.WhenTheLpaCanBeUsed
		lpa.WhenTheLpaCanBeUsedIsDefault = false
	}

	lpa.HowAttorneysMakeDecisionsDetails = c.HowAttorneysMakeDecisionsDetails
	lpa.HowReplacementAttorneysStepIn = c.HowReplacementAttorneysStepIn
	lpa.HowReplacementAttorneysStepInDetails = c.HowReplacementAttorneysStepInDetails
	lpa.HowReplacementAttorneysMakeDecisionsDetails = c.HowReplacementAttorneysMakeDecisionsDetails

	return nil
}

type AuthorisedSignatoryPreRegistrationCorrection struct {
	shared.AuthorisedSignatoryCorrection
}

func (c AuthorisedSignatoryPreRegistrationCorrection) Apply(lpa *shared.Lpa) []shared.FieldError {
	if c.FirstNames != "" && c.LastName != "" {
		as := &shared.AuthorisedSignatory{
			Person: shared.Person{
				FirstNames: c.FirstNames,
				LastName:   c.LastName,
			},
		}

		lpa.AuthorisedSignatory = as
	}

	return nil
}

type IndependentWitnessPreRegistrationCorrection struct {
	shared.IndependentWitnessCorrection
}

func (c IndependentWitnessPreRegistrationCorrection) Apply(lpa *shared.Lpa) []shared.FieldError {
	if c.FirstNames != "" && c.LastName != "" {
		lpa.IndependentWitness.Person = shared.Person{
			FirstNames: c.FirstNames,
			LastName:   c.LastName,
		}
	}
	if c.Phone != "" {
		lpa.IndependentWitness.Phone = c.Phone
	}

	if !c.Address.IsZero() {
		lpa.IndependentWitness.Address = c.Address
	}

	return nil
}

type WitnessedByPreRegistrationCorrection struct {
	shared.WitnessedByCorrection
}

func (c WitnessedByPreRegistrationCorrection) Apply(lpa *shared.Lpa) []shared.FieldError {
	if !c.WitnessedByCertificateProviderAt.IsZero() {
		lpa.WitnessedByCertificateProviderAt = c.WitnessedByCertificateProviderAt
	}

	if !c.WitnessedByIndependentWitnessAt.IsZero() {
		lpa.WitnessedByIndependentWitnessAt = &c.WitnessedByIndependentWitnessAt
	}

	return nil
}

func (c Correction) Apply(lpa *shared.Lpa) []shared.FieldError {
	isInvalidOnlineLPASignedAtChange := !c.SignedAt.IsZero() && !c.SignedAt.Equal(lpa.SignedAt) &&
		lpa.Channel == shared.ChannelOnline

	if isInvalidOnlineLPASignedAtChange {
		return []shared.FieldError{{Source: signedAt, Detail: "LPA Signed on date cannot be changed for online LPAs"}}
	}

	if lpa.Status == shared.LpaStatusRegistered {
		return []shared.FieldError{{Source: "/type", Detail: "Cannot make corrections to a Registered LPA"}}
	}

	if fieldErrors := c.Donor.Apply(lpa); len(fieldErrors) > 0 {
		return fieldErrors
	}

	if fieldErrors := c.CertificateProvider.Apply(lpa); len(fieldErrors) > 0 {
		return fieldErrors
	}

	if fieldErrors := c.Attorney.Apply(lpa); len(fieldErrors) > 0 {
		return fieldErrors
	}

	if fieldErrors := c.TrustCorporation.Apply(lpa); len(fieldErrors) > 0 {
		return fieldErrors
	}

	if fieldErrors := c.AttorneyAppointmentType.Apply(lpa); len(fieldErrors) > 0 {
		return fieldErrors
	}

	if fieldErrors := c.AuthorisedSignatory.Apply(lpa); len(fieldErrors) > 0 {
		return fieldErrors
	}

	if fieldErrors := c.IndependentWitness.Apply(lpa); len(fieldErrors) > 0 {
		return fieldErrors
	}

	if fieldErrors := c.WitnessedBy.Apply(lpa); len(fieldErrors) > 0 {
		return fieldErrors
	}

	lpa.SignedAt = c.SignedAt

	return nil
}

func validateCorrection(changes []shared.Change, lpa *shared.Lpa) (Correction, []shared.FieldError) {
	var data Correction

	data.SignedAt = lpa.SignedAt
	data.AttorneyAppointmentType.HowReplacementAttorneysStepIn = lpa.HowReplacementAttorneysStepIn
	data.AttorneyAppointmentType.HowReplacementAttorneysStepInDetails = lpa.HowReplacementAttorneysStepInDetails

	data.Donor.FirstNames = lpa.Donor.FirstNames
	data.Donor.LastName = lpa.Donor.LastName
	data.Donor.OtherNamesKnownBy = lpa.Donor.OtherNamesKnownBy
	data.Donor.DateOfBirth = lpa.Donor.DateOfBirth
	data.Donor.Address = lpa.Donor.Address
	data.Donor.Email = lpa.Donor.Email

	data.CertificateProvider.FirstNames = lpa.CertificateProvider.FirstNames
	data.CertificateProvider.LastName = lpa.CertificateProvider.LastName
	data.CertificateProvider.Address = lpa.CertificateProvider.Address
	data.CertificateProvider.Email = lpa.CertificateProvider.Email
	data.CertificateProvider.Phone = lpa.CertificateProvider.Phone
	if lpa.CertificateProvider.SignedAt != nil {
		data.CertificateProvider.SignedAt = *lpa.CertificateProvider.SignedAt
	}

	if lpa.AuthorisedSignatory != nil {
		data.AuthorisedSignatory.FirstNames = lpa.AuthorisedSignatory.FirstNames
		data.AuthorisedSignatory.LastName = lpa.AuthorisedSignatory.LastName
	}

	if lpa.IndependentWitness != nil {
		data.IndependentWitness.FirstNames = lpa.IndependentWitness.FirstNames
		data.IndependentWitness.LastName = lpa.IndependentWitness.LastName
		data.IndependentWitness.Address = lpa.IndependentWitness.Address
	}

	data.WitnessedBy.WitnessedByCertificateProviderAt = lpa.WitnessedByCertificateProviderAt
	if lpa.WitnessedByIndependentWitnessAt != nil {
		data.WitnessedBy.WitnessedByIndependentWitnessAt = *lpa.WitnessedByIndependentWitnessAt
	}

	parser := parse.Changes(changes).
		Field(signedAt, &data.SignedAt, parse.Validate(validate.NotEmpty()), parse.Optional()).
		Prefix("/donor", validateDonor(&data.Donor), parse.Optional()).
		Prefix("/certificateProvider", validateCertificateProvider(&data.CertificateProvider), parse.Optional()).
		Prefix("/attorneys", func(p *parse.Parser) []shared.FieldError {
			return p.
				EachKey(func(key string, p *parse.Parser) []shared.FieldError {
					attorneyIdx, ok := lpa.FindAttorneyIndex(key)

					if !ok || (data.Attorney.Index != nil && *data.Attorney.Index != attorneyIdx) {
						return p.OutOfRange()
					}

					data.Attorney.Index = &attorneyIdx
					data.Attorney.FirstNames = lpa.Attorneys[attorneyIdx].FirstNames
					data.Attorney.LastName = lpa.Attorneys[attorneyIdx].LastName
					data.Attorney.DateOfBirth = lpa.Attorneys[attorneyIdx].DateOfBirth
					data.Attorney.Address = lpa.Attorneys[attorneyIdx].Address
					data.Attorney.Email = lpa.Attorneys[attorneyIdx].Email
					data.Attorney.Mobile = lpa.Attorneys[attorneyIdx].Mobile

					if lpa.Attorneys[attorneyIdx].SignedAt != nil {
						data.Attorney.SignedAt = *lpa.Attorneys[attorneyIdx].SignedAt
					}

					return validateAttorney(&data.Attorney, p)
				}).
				Consumed()
		}, parse.Optional()).
		Prefix("/trustCorporations", func(p *parse.Parser) []shared.FieldError {
			return p.
				EachKey(func(key string, p *parse.Parser) []shared.FieldError {
					i, ok := lpa.FindTrustCorporationIndex(key)

					if !ok || (data.TrustCorporation.Index != nil && *data.TrustCorporation.Index != i) {
						return p.OutOfRange()
					}

					data.TrustCorporation.Index = &i
					data.TrustCorporation.Name = lpa.TrustCorporations[i].Name
					data.TrustCorporation.CompanyNumber = lpa.TrustCorporations[i].CompanyNumber
					data.TrustCorporation.Email = lpa.TrustCorporations[i].Email
					data.TrustCorporation.Address = lpa.TrustCorporations[i].Address
					data.TrustCorporation.Mobile = lpa.TrustCorporations[i].Mobile
					data.TrustCorporation.Signatories = lpa.TrustCorporations[i].Signatories[:]

					return validateTrustCorporation(&data.TrustCorporation, p)
				}).
				Consumed()
		}, parse.Optional()).
		Prefix("/authorisedSignatory", validateAuthorisedSignatory(&data.AuthorisedSignatory), parse.Optional()).
		Prefix("/independentWitness", validateIndependentWitness(&data.IndependentWitness), parse.Optional()).
		Field("/witnessedByCertificateProviderAt", &data.WitnessedBy.WitnessedByCertificateProviderAt, parse.Validate(validate.NotEmpty()), parse.Optional()).
		Field("/witnessedByIndependentWitnessAt", &data.WitnessedBy.WitnessedByIndependentWitnessAt, parse.Validate(validate.NotEmpty()), parse.Optional())

	activeAttorneyCount, replacementAttorneyCount := shared.CountAttorneys(lpa.Attorneys, lpa.TrustCorporations)

	if activeAttorneyCount > 1 {
		parser.Field("/howAttorneysMakeDecisions", &data.AttorneyAppointmentType.HowAttorneysMakeDecisions,
			parse.Old(&lpa.HowAttorneysMakeDecisions),
			parse.Validate(validate.Valid()),
			parse.Optional())
	}

	attorneysJointlyForSomeSeverallyForOthers := data.AttorneyAppointmentType.HowAttorneysMakeDecisions == shared.HowMakeDecisionsJointlyForSomeSeverallyForOthers ||
		data.AttorneyAppointmentType.HowAttorneysMakeDecisions.Unset() && lpa.HowAttorneysMakeDecisions == shared.HowMakeDecisionsJointlyForSomeSeverallyForOthers

	if attorneysJointlyForSomeSeverallyForOthers {
		parser.Field("/howAttorneysMakeDecisionsDetails", &data.AttorneyAppointmentType.HowAttorneysMakeDecisionsDetails,
			parse.Old(&lpa.HowAttorneysMakeDecisionsDetails),
			parse.Validate(validate.NotEmpty()))
	} else {
		parser.Field("/howAttorneysMakeDecisionsDetails", &data.AttorneyAppointmentType.HowAttorneysMakeDecisionsDetails,
			parse.Old(&lpa.HowAttorneysMakeDecisionsDetails),
			parse.Validate(validate.Empty()),
			parse.Optional())
	}

	attorneysJointlyAndSeverally := data.AttorneyAppointmentType.HowAttorneysMakeDecisions == shared.HowMakeDecisionsJointlyAndSeverally ||
		data.AttorneyAppointmentType.HowAttorneysMakeDecisions.Unset() && lpa.HowAttorneysMakeDecisions == shared.HowMakeDecisionsJointlyAndSeverally

	if replacementAttorneyCount > 0 && attorneysJointlyAndSeverally {
		parser.Field("/howReplacementAttorneysStepIn", &data.AttorneyAppointmentType.HowReplacementAttorneysStepIn,
			parse.Validate(validate.Valid()),
			parse.Optional())
	}

	if data.AttorneyAppointmentType.HowReplacementAttorneysStepIn == shared.HowStepInAnotherWay {
		parser.Field("/howReplacementAttorneysStepInDetails", &data.AttorneyAppointmentType.HowReplacementAttorneysStepInDetails,
			parse.Validate(validate.NotEmpty()))
	} else {
		parser.Field("/howReplacementAttorneysStepInDetails", &data.AttorneyAppointmentType.HowReplacementAttorneysStepInDetails,
			parse.Validate(validate.Empty()),
			parse.Optional())
	}

	if replacementAttorneyCount > 1 && (data.AttorneyAppointmentType.HowReplacementAttorneysStepIn == shared.HowStepInAllCanNoLongerAct || !attorneysJointlyAndSeverally) {
		parser.Field("/howReplacementAttorneysMakeDecisions", &data.AttorneyAppointmentType.HowReplacementAttorneysMakeDecisions,
			parse.Old(&lpa.HowReplacementAttorneysMakeDecisions),
			parse.Validate(validate.Valid()),
			parse.Optional())
	}

	replacementsJointlyForSomeSeverallyForOthers := data.AttorneyAppointmentType.HowReplacementAttorneysMakeDecisions == shared.HowMakeDecisionsJointlyForSomeSeverallyForOthers ||
		data.AttorneyAppointmentType.HowReplacementAttorneysMakeDecisions.Unset() && lpa.HowReplacementAttorneysMakeDecisions == shared.HowMakeDecisionsJointlyForSomeSeverallyForOthers

	if replacementsJointlyForSomeSeverallyForOthers {
		parser.Field("/howReplacementAttorneysMakeDecisionsDetails", &data.AttorneyAppointmentType.HowReplacementAttorneysMakeDecisionsDetails,
			parse.Old(&lpa.HowReplacementAttorneysMakeDecisionsDetails),
			parse.Validate(validate.NotEmpty()))
	} else {
		parser.Field("/howReplacementAttorneysMakeDecisionsDetails", &data.AttorneyAppointmentType.HowReplacementAttorneysMakeDecisionsDetails,
			parse.Old(&lpa.HowReplacementAttorneysMakeDecisionsDetails),
			parse.Validate(validate.Empty()),
			parse.Optional())
	}

	if lpa.LpaType == shared.LpaTypePersonalWelfare {
		parser.Field("/lifeSustainingTreatmentOption", &data.AttorneyAppointmentType.LifeSustainingTreatmentOption,
			parse.Old(&lpa.LifeSustainingTreatmentOption),
			parse.Validate(validate.Valid()),
			parse.Optional())
	}

	if lpa.LpaType == shared.LpaTypePropertyAndAffairs {
		parser.Field("/whenTheLpaCanBeUsed", &data.AttorneyAppointmentType.WhenTheLpaCanBeUsed,
			parse.Old(&lpa.WhenTheLpaCanBeUsed),
			parse.Validate(validate.Valid()),
			parse.Optional())
	}

	errors := parser.Consumed()

	return data, errors
}

func validateAttorney(attorney *AttorneyPreRegistrationCorrection, p *parse.Parser) []shared.FieldError {
	return p.
		Field("/firstNames", &attorney.FirstNames, parse.Validate(validate.NotEmpty()), parse.Optional()).
		Field("/lastName", &attorney.LastName, parse.Validate(validate.NotEmpty()), parse.Optional()).
		Field("/dateOfBirth", &attorney.DateOfBirth, parse.Validate(validate.Date()), parse.Optional()).
		Field("/email", &attorney.Email, parse.Optional()).
		Field("/mobile", &attorney.Mobile, parse.Optional()).
		Prefix("/address", validateAddress(&attorney.Address), parse.Optional()).
		Field(signedAt, &attorney.SignedAt, parse.Validate(validate.NotEmpty()), parse.Optional()).
		Consumed()
}

func validateDonor(donor *DonorPreRegistrationCorrection) func(p *parse.Parser) []shared.FieldError {
	return func(p *parse.Parser) []shared.FieldError {
		return p.
			Field("/firstNames", &donor.FirstNames, parse.Validate(validate.NotEmpty()), parse.Optional()).
			Field("/lastName", &donor.LastName, parse.Validate(validate.NotEmpty()), parse.Optional()).
			Field("/otherNamesKnownBy", &donor.OtherNamesKnownBy, parse.Optional()).
			Field("/dateOfBirth", &donor.DateOfBirth, parse.Validate(validate.Date()), parse.Optional()).
			Prefix("/address", validateAddress(&donor.Address), parse.Optional()).
			Field("/email", &donor.Email, parse.Optional()).
			Consumed()
	}
}

func validateCertificateProvider(certificateProvider *CertificateProviderPreRegistrationCorrection) func(p *parse.Parser) []shared.FieldError {
	return func(p *parse.Parser) []shared.FieldError {
		return p.
			Field("/firstNames", &certificateProvider.FirstNames, parse.Validate(validate.NotEmpty()), parse.Optional()).
			Field("/lastName", &certificateProvider.LastName, parse.Validate(validate.NotEmpty()), parse.Optional()).
			Prefix("/address", validateAddress(&certificateProvider.Address), parse.Optional()).
			Field("/email", &certificateProvider.Email, parse.Optional()).
			Field("/phone", &certificateProvider.Phone, parse.Optional()).
			Field("/signedAt", &certificateProvider.SignedAt, parse.Optional()).
			Consumed()
	}
}

func validateTrustCorporation(trustCorporation *TrustCorporationPreRegistrationCorrection, p *parse.Parser) []shared.FieldError {
	return p.
		Field("/name", &trustCorporation.Name, parse.Validate(validate.NotEmpty()), parse.Optional()).
		Field("/companyNumber", &trustCorporation.CompanyNumber, parse.Validate(validate.NotEmpty()), parse.Optional()).
		Field("/email", &trustCorporation.Email, parse.Optional()).
		Prefix("/address", validateAddress(&trustCorporation.Address), parse.Optional()).
		Field("/mobile", &trustCorporation.Mobile, parse.Optional()).
		Prefix("/signatories", func(p *parse.Parser) []shared.FieldError {
			return p.
				Each(func(i int, p *parse.Parser) []shared.FieldError {
					if i > 1 {
						return p.OutOfRange()
					}

					return validateSignatory(&trustCorporation.Signatories[i], p)
				}).
				Consumed()
		}, parse.Optional()).
		Consumed()
}

func validateAuthorisedSignatory(as *AuthorisedSignatoryPreRegistrationCorrection) func(p *parse.Parser) []shared.FieldError {
	return func(p *parse.Parser) []shared.FieldError {
		return p.
			Field("/firstNames", &as.FirstNames, parse.Validate(validate.NotEmpty()), parse.Optional()).
			Field("/lastName", &as.LastName, parse.Validate(validate.NotEmpty()), parse.Optional()).
			Consumed()
	}
}

func validateIndependentWitness(iw *IndependentWitnessPreRegistrationCorrection) func(p *parse.Parser) []shared.FieldError {
	return func(p *parse.Parser) []shared.FieldError {
		return p.
			Field("/firstNames", &iw.FirstNames, parse.Validate(validate.NotEmpty()), parse.Optional()).
			Field("/lastName", &iw.LastName, parse.Validate(validate.NotEmpty()), parse.Optional()).
			Field("/phone", &iw.Phone, parse.Validate(validate.NotEmpty()), parse.Optional()).
			Prefix("/address", validateAddress(&iw.Address), parse.Optional()).
			Consumed()
	}
}

func validateAddress(address *shared.Address) func(p *parse.Parser) []shared.FieldError {
	return func(p *parse.Parser) []shared.FieldError {
		return p.
			Field("/line1", &address.Line1, parse.Optional()).
			Field("/line2", &address.Line2, parse.Optional()).
			Field("/line3", &address.Line3, parse.Optional()).
			Field("/town", &address.Town, parse.Optional()).
			Field("/postcode", &address.Postcode, parse.Optional()).
			Field("/country", &address.Country, parse.Validate(validate.Country()), parse.Optional()).
			Consumed()
	}
}

func validateSignatory(signatory *shared.Signatory, p *parse.Parser) []shared.FieldError {
	return p.
		Field("/firstNames", &signatory.FirstNames, parse.Validate(validate.NotEmpty()), parse.Optional()).
		Field("/lastName", &signatory.LastName, parse.Validate(validate.NotEmpty()), parse.Optional()).
		Field("/professionalTitle", &signatory.ProfessionalTitle, parse.Optional()).
		Consumed()
}
