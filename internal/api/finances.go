package api

import "net/url"

type Debt struct {
	DebtID          string  `json:"debt_id"`
	Name            string  `json:"name"`
	Type            string  `json:"type,omitempty"`
	Balance         float64 `json:"balance,omitempty"`
	APR             float64 `json:"apr,omitempty"`
	MonthlyPayment  float64 `json:"monthly_payment,omitempty"`
	OriginalBalance float64 `json:"original_balance,omitempty"`
	Notes           string  `json:"notes,omitempty"`
	PayoffMonths    int     `json:"payoff_months,omitempty"`
	TotalInterest   float64 `json:"total_interest,omitempty"`
}

type DebtInput struct {
	Name            string  `json:"name,omitempty"`
	Type            string  `json:"type,omitempty"`
	Balance         float64 `json:"balance,omitempty"`
	APR             float64 `json:"apr,omitempty"`
	MonthlyPayment  float64 `json:"monthly_payment,omitempty"`
	OriginalBalance float64 `json:"original_balance,omitempty"`
	Notes           string  `json:"notes,omitempty"`
}

type Income struct {
	IncomeID      string  `json:"income_id"`
	Name          string  `json:"name"`
	Amount        float64 `json:"amount,omitempty"`
	Frequency     string  `json:"frequency,omitempty"`
	MonthlyAmount float64 `json:"monthly_amount,omitempty"`
	Notes         string  `json:"notes,omitempty"`
	StartDate     string  `json:"start_date,omitempty"`
}

type IncomeInput struct {
	Name      string  `json:"name,omitempty"`
	Amount    float64 `json:"amount,omitempty"`
	Frequency string  `json:"frequency,omitempty"`
	Notes     string  `json:"notes,omitempty"`
	StartDate string  `json:"start_date,omitempty"`
}

type FixedExpense struct {
	ExpenseID     string  `json:"expense_id"`
	Name          string  `json:"name,omitempty"`
	Category      string  `json:"category,omitempty"`
	Amount        float64 `json:"amount,omitempty"`
	Frequency     string  `json:"frequency,omitempty"`
	MonthlyAmount float64 `json:"monthly_amount,omitempty"`
	DueDay        int     `json:"due_day,omitempty"`
	Notes         string  `json:"notes,omitempty"`
}

type FixedExpenseInput struct {
	Name      string  `json:"name,omitempty"`
	Category  string  `json:"category,omitempty"`
	Amount    float64 `json:"amount,omitempty"`
	Frequency string  `json:"frequency,omitempty"`
	DueDay    int     `json:"due_day,omitempty"`
	Notes     string  `json:"notes,omitempty"`
}

type FinancesSummary struct {
	TotalIncome   float64 `json:"total_income,omitempty"`
	TotalExpenses float64 `json:"total_expenses,omitempty"`
	TotalDebt     float64 `json:"total_debt,omitempty"`
	NetCashFlow   float64 `json:"net_cash_flow,omitempty"`
	NetWorth      float64 `json:"net_worth,omitempty"`
}

func (c *Client) ListDebts() ([]Debt, error)            { return GetList[Debt](c, "/debts") }
func (c *Client) CreateDebt(in DebtInput) (Debt, error) { return PostOne[Debt](c, "/debts", in) }
func (c *Client) UpdateDebt(id string, in DebtInput) (Debt, error) {
	return PutOne[Debt](c, "/debts/"+url.PathEscape(id), in)
}
func (c *Client) DeleteDebt(id string) error { return c.Delete("/debts/" + url.PathEscape(id)) }

func (c *Client) ListIncome() ([]Income, error) { return GetList[Income](c, "/income") }
func (c *Client) CreateIncome(in IncomeInput) (Income, error) {
	return PostOne[Income](c, "/income", in)
}
func (c *Client) UpdateIncome(id string, in IncomeInput) (Income, error) {
	return PutOne[Income](c, "/income/"+url.PathEscape(id), in)
}
func (c *Client) DeleteIncome(id string) error { return c.Delete("/income/" + url.PathEscape(id)) }

func (c *Client) ListFixedExpenses() ([]FixedExpense, error) {
	return GetList[FixedExpense](c, "/fixed-expenses")
}
func (c *Client) CreateFixedExpense(in FixedExpenseInput) (FixedExpense, error) {
	return PostOne[FixedExpense](c, "/fixed-expenses", in)
}
func (c *Client) UpdateFixedExpense(id string, in FixedExpenseInput) (FixedExpense, error) {
	return PutOne[FixedExpense](c, "/fixed-expenses/"+url.PathEscape(id), in)
}
func (c *Client) DeleteFixedExpense(id string) error {
	return c.Delete("/fixed-expenses/" + url.PathEscape(id))
}

func (c *Client) FinancesSummary() (FinancesSummary, error) {
	return GetOne[FinancesSummary](c, "/finances/summary")
}
