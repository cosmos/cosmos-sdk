package ecdsa

func (suite *EcdsaSuite) TestSkString() {
	suite.Require().Equal("secp256r1{-}", suite.sk.String())
}

func (suite *EcdsaSuite) xTestSkEqual() {
	require := suite.Require()

	skOther, err := GenSecp256r1()
	require.NoError(err)
	require.False(suite.sk.Equals(skOther))

	skOther2 := ecdsaSK{skOther.(ecdsaSK).PrivateKey}
	require.True(skOther.Equals(skOther2))
	require.True(skOther2.Equals(skOther), "Equals must be reflexive")
}
