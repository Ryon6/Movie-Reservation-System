package repository

// isForeignKeyConstraintError 检查错误是否是MySQL外键约束错误（错误码1451）
// func isForeignKeyConstraintError(err error) bool {
// 	var mysqlErr *mysql.MySQLError
// 	if errors.As(err, &mysqlErr) {
// 		return mysqlErr.Number == 1451
// 	}
// 	return false
// }
