package keeper_test

import (
	"fmt"

	"cosmossdk.io/x/nft"
)

var (
	ExpClass = nft.Class{
		Id:          testClassID,
		Name:        testClassName,
		Symbol:      testClassSymbol,
		Description: testClassDescription,
		Uri:         testClassURI,
		UriHash:     testClassURIHash,
	}

	ExpNFT = nft.NFT{
		ClassId: testClassID,
		Id:      testID,
		Uri:     testURI,
	}
)

func (s *TestSuite) TestSend() {
	err := s.nftKeeper.SaveClass(s.ctx, ExpClass)
	s.Require().NoError(err)

	actual, has := s.nftKeeper.GetClass(s.ctx, testClassID)
	s.Require().True(has)
	s.Require().EqualValues(ExpClass, actual)

	err = s.nftKeeper.Mint(s.ctx, ExpNFT, s.addrs[0])
	s.Require().NoError(err)

	expGenesis := &nft.GenesisState{
		Classes: []*nft.Class{&ExpClass},
		Entries: []*nft.Entry{{
			Owner: s.encodedAddrs[0],
			Nfts:  []*nft.NFT{&ExpNFT},
		}},
	}
	genesis, err := s.nftKeeper.ExportGenesis(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal(expGenesis, genesis)

	testCases := []struct {
		name   string
		req    *nft.MsgSend
		expErr bool
		errMsg string
	}{
		{
			name: "empty nft id",
			req: &nft.MsgSend{
				ClassId:  testClassID,
				Id:       "",
				Sender:   s.encodedAddrs[0],
				Receiver: s.encodedAddrs[1],
			},
			expErr: true,
			errMsg: "empty nft id",
		},
		{
			name: "empty class id",
			req: &nft.MsgSend{
				ClassId:  "",
				Id:       testID,
				Sender:   s.encodedAddrs[0],
				Receiver: s.encodedAddrs[1],
			},
			expErr: true,
			errMsg: "empty class id",
		},
		{
			name: "invalid class id",
			req: &nft.MsgSend{
				ClassId:  "invalid ClassId",
				Id:       testID,
				Sender:   s.encodedAddrs[0],
				Receiver: s.encodedAddrs[1],
			},
			expErr: true,
			errMsg: "unauthorized",
		},
		{
			name: "invalid nft id",
			req: &nft.MsgSend{
				ClassId:  testClassID,
				Id:       "invalid Id",
				Sender:   s.encodedAddrs[0],
				Receiver: s.encodedAddrs[1],
			},
			expErr: true,
			errMsg: "unauthorized",
		},
		{
			name: "unauthorized sender",
			req: &nft.MsgSend{
				ClassId:  testClassID,
				Id:       testID,
				Sender:   s.encodedAddrs[1],
				Receiver: s.encodedAddrs[2],
			},
			expErr: true,
			errMsg: fmt.Sprintf("%s is not the owner of nft %s", s.encodedAddrs[1], testID),
		},
		{
			name: "valid transaction",
			req: &nft.MsgSend{
				ClassId:  testClassID,
				Id:       testID,
				Sender:   s.encodedAddrs[0],
				Receiver: s.encodedAddrs[1],
			},
			expErr: false,
			errMsg: "",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			_, err := s.nftKeeper.Send(s.ctx, tc.req)
			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errMsg)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func (s *TestSuite) TestMsgNewClass() {
	msg := &nft.MsgNewClass{
		ClassId:     testClassID,
		Name:        testClassName,
		Symbol:      testClassSymbol,
		Description: testClassDescription,
		Uri:         testClassURI,
		UriHash:     testClassURIHash,
	}

	_, err := s.nftKeeper.MsgNewClass(s.ctx, msg)
	s.Require().NoError(err)

	actual, has := s.nftKeeper.GetClass(s.ctx, testClassID)
	s.Require().True(has)
	s.Require().EqualValues(nft.Class{
		Id:          testClassID,
		Name:        testClassName,
		Symbol:      testClassSymbol,
		Description: testClassDescription,
		Uri:         testClassURI,
		UriHash:     testClassURIHash,
	}, actual)
}

func (s *TestSuite) TestMsgUpdateClass() {
	class := nft.Class{
		Id:          testClassID,
		Name:        testClassName,
		Symbol:      testClassSymbol,
		Description: testClassDescription,
		Uri:         testClassURI,
		UriHash:     testClassURIHash,
	}
	err := s.nftKeeper.SaveClass(s.ctx, class)
	s.Require().NoError(err)

	msg := &nft.MsgUpdateClass{
		ClassId:     testClassID,
		Name:        "Updated Name",
		Symbol:      "Updated Symbol",
		Description: "Updated Description",
		Uri:         "Updated URI",
		UriHash:     "Updated URI Hash",
	}

	_, err = s.nftKeeper.MsgUpdateClass(s.ctx, msg)
	s.Require().NoError(err)

	actual, has := s.nftKeeper.GetClass(s.ctx, testClassID)
	s.Require().True(has)
	s.Require().EqualValues(nft.Class{
		Id:          testClassID,
		Name:        "Updated Name",
		Symbol:      "Updated Symbol",
		Description: "Updated Description",
		Uri:         "Updated URI",
		UriHash:     "Updated URI Hash",
	}, actual)
}

func (s *TestSuite) TestMsgMintNFT() {
	class := nft.Class{
		Id:          testClassID,
		Name:        testClassName,
		Symbol:      testClassSymbol,
		Description: testClassDescription,
		Uri:         testClassURI,
		UriHash:     testClassURIHash,
	}
	err := s.nftKeeper.SaveClass(s.ctx, class)
	s.Require().NoError(err)

	msg := &nft.MsgMintNFT{
		ClassId:  testClassID,
		Id:       testID,
		Uri:      testURI,
		UriHash:  testURIHash,
		Receiver: s.encodedAddrs[0],
	}

	_, err = s.nftKeeper.MsgMintNFT(s.ctx, msg)
	s.Require().NoError(err)

	actual, has := s.nftKeeper.GetNFT(s.ctx, testClassID, testID)
	s.Require().True(has)
	s.Require().EqualValues(nft.NFT{
		ClassId: testClassID,
		Id:      testID,
		Uri:     testURI,
		UriHash: testURIHash,
	}, actual)
}

func (s *TestSuite) TestMsgBurnNFT() {
	class := nft.Class{
		Id:          testClassID,
		Name:        testClassName,
		Symbol:      testClassSymbol,
		Description: testClassDescription,
		Uri:         testClassURI,
		UriHash:     testClassURIHash,
	}
	err := s.nftKeeper.SaveClass(s.ctx, class)
	s.Require().NoError(err)

	nft := nft.NFT{
		ClassId: testClassID,
		Id:      testID,
		Uri:     testURI,
		UriHash: testURIHash,
	}
	err = s.nftKeeper.Mint(s.ctx, nft, s.addrs[0])
	s.Require().NoError(err)

	msg := &nft.MsgBurnNFT{
		ClassId: testClassID,
		Id:      testID,
	}

	_, err = s.nftKeeper.MsgBurnNFT(s.ctx, msg)
	s.Require().NoError(err)

	_, has := s.nftKeeper.GetNFT(s.ctx, testClassID, testID)
	s.Require().False(has)
}

func (s *TestSuite) TestMsgUpdateNFT() {
	class := nft.Class{
		Id:          testClassID,
		Name:        testClassName,
		Symbol:      testClassSymbol,
		Description: testClassDescription,
		Uri:         testClassURI,
		UriHash:     testClassURIHash,
	}
	err := s.nftKeeper.SaveClass(s.ctx, class)
	s.Require().NoError(err)

	nft := nft.NFT{
		ClassId: testClassID,
		Id:      testID,
		Uri:     testURI,
		UriHash: testURIHash,
	}
	err = s.nftKeeper.Mint(s.ctx, nft, s.addrs[0])
	s.Require().NoError(err)

	msg := &nft.MsgUpdateNFT{
		ClassId: testClassID,
		Id:      testID,
		Uri:     "Updated URI",
		UriHash: "Updated URI Hash",
	}

	_, err = s.nftKeeper.MsgUpdateNFT(s.ctx, msg)
	s.Require().NoError(err)

	actual, has := s.nftKeeper.GetNFT(s.ctx, testClassID, testID)
	s.Require().True(has)
	s.Require().EqualValues(nft.NFT{
		ClassId: testClassID,
		Id:      testID,
		Uri:     "Updated URI",
		UriHash: "Updated URI Hash",
	}, actual)
}
