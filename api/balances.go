package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/triplan-planning/api-go/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// @Summary      Returns the balance of every user in the group
// @Accept       json
// @Param        id   path      string  true  "Group ID"
// @Success      200  {object}  model.Transaction
// @Router       /groups/{id}/balances [get]
func (api *Api) GetGroupBalances(c *fiber.Ctx) error {

	groupId, err := getId(c.Params("id"))
	if err != nil {
		return err
	}

	// Step 1 : get group users
	groupRaw := api.groupsColl.FindOne(
		c.Context(),
		bson.M{"_id": groupId},
	)
	if groupRaw.Err() != nil {
		return groupRaw.Err()
	}
	var group model.Group
	err = groupRaw.Decode(&group)
	if err != nil {
		return err
	}
	// Step 1 end : users stored in group.users var

	// Step 2 get all transactions in group
	transactionsRaw, err := api.transactionsColl.Find(
		c.Context(),
		model.Transaction{Group: groupId},
	)
	if err != nil {
		return err
	}

	var transactions []model.Transaction
	err = transactionsRaw.All(c.Context(), &transactions)
	if err != nil {
		return err
	}

	balanceMap := make(map[primitive.ObjectID]*model.Balance)
	for _, userId := range group.Users {
		balanceMap[userId] = &model.Balance{
			PositiveAmount: 0,
			NegativeAmount: 0,
			TotalAmount:    0,
		}
	}

	for _, transaction := range transactions {
		err = transaction.ComputePrices()
		if err != nil {
			return err
		}
		payer := transaction.PaidBy
		balanceMap[payer].PositiveAmount += transaction.Amount

		for _, target := range transaction.PaidFor {
			balanceMap[target.User].NegativeAmount += target.ComputedPrice
		}

	}

	for _, balance := range balanceMap {
		balance.TotalAmount = int32(
			balance.PositiveAmount - balance.NegativeAmount,
		)
	}

	return c.JSON(balanceMap)
}
