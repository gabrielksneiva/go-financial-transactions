package services

import (
	d "github.com/financialkafkaconsumerproject/producer/domain"
)

// Struct principal da service
type StatementService struct {
	Repo        d.TransactionRepository
	BalanceRepo d.BalanceRepository
}

// Construtor
func NewStatementService(r d.TransactionRepository, b d.BalanceRepository) *StatementService {
	return &StatementService{
		Repo:        r,
		BalanceRepo: b,
	}
}

// Retorna saldo
func (s *StatementService) GetBalance(userID string) (float64, error) {
	balance, err := s.BalanceRepo.GetBalance(userID)
	if err != nil {
		return 0.0, err
	}
	return balance.Amount, nil
}

// Retorna transações
func (s *StatementService) GetTransactions(userID string) ([]d.Transaction, error) {
	return s.Repo.GetByUser(userID)
}

// Struct para retorno unificado de extrato
type Statement struct {
	Balance      float64
	Transactions []d.Transaction
}

// Retorna extrato completo (saldo + transações)
func (s *StatementService) GetStatement(userID string) (*Statement, error) {
	balance, err := s.GetBalance(userID)
	if err != nil {
		return nil, err
	}

	transactions, err := s.GetTransactions(userID)
	if err != nil {
		return nil, err
	}

	return &Statement{
		Balance:      balance,
		Transactions: transactions,
	}, nil
}
