package main

import (
	"BootsDB/storage_manager"
	"bufio"
	"fmt"
	"os"
	"strings"
)

type QueryType string

const (
	INSERT QueryType = "INSERT"
	SELECT QueryType = "SELECT"
)

func scanQuery(input string) QueryType {
	cmd := strings.ToUpper(strings.TrimSpace(input))
	if strings.HasPrefix(cmd, string(INSERT)) {
		return INSERT
	}
	if strings.HasPrefix(cmd, string(SELECT)) {
		return SELECT
	}
	return ""
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	storage_engine := storage_manager.InitializeStorage("Boots.db")
	pager := storage_manager.InitializePager(storage_engine)
	btree := storage_manager.InitializeBtree(pager)
	for {
		fmt.Print("BootsDB> ")
		scanner.Scan()
		input := scanner.Text()

		if input == ".exit" {
			pager.FlushCache()
			break
		}

		switch scanQuery(input) {
		case INSERT:
			parts := strings.Fields(input)
			username := parts[1]
			email := parts[2]
			btree.Insert(username, email)

		case SELECT:
			btree.Select()

		default:
			fmt.Println("Unknown command")
		}
	}
}

//DONE:
//Work on select functionality: DONE
//Work on marking pages dirty when we change cached pages DONE
//Work on writing back to disk DONE
//Ensure we have correct functionality for when root gets full DONE
//Ensure tests work DONE

//TODO:
//Implement simple query parse
//	-Lets implement this so it can read from command line and file
//	-The reason I am doing this is to create a clean interface between parse/b+ tree since this interface
//	 may change some parts of the b+ tree
//	-Example
// Query: SELECT customer_name, SUM(order_total) 
//        FROM customers JOIN orders ON customers.id = orders.customer_id
//        WHERE order_date > '2023-01-01'
//        GROUP BY customer_name
//        HAVING SUM(order_total) > 1000
//        ORDER BY SUM(order_total) DESC
//
// 					Final Result
//                            ↑
//                      Sort (by total DESC)
//                            ↑
//                    Filter (total > 1000)
//                            ↑
//               Aggregation (GROUP BY customer_name)
//                            ↑
//                      Hash Join
//                       ↗     ↖
//        Table Scan (customers)  Filter (date > '2023-01-01')
//                                       ↑
//                               Table Scan (orders)
//
// High Level Overview of Query Execution:
// SQL query execution involves parsing a query into a syntax 
// tree(Do we have to transform to tree first or can we just create DAG first), transforming it into a
// logical plan, optimizing it into a physical Directed Acyclic Graph (DAG), and then 
// executing this DAG by flowing data through each node where specific operations transform
// the data until it reaches the final node, producing the requested result.

//Implement composite index
//	-This is when we have multiple keys for an index
//Handle duplicate indexes(Approach: Append recordId - watch cmu B+ Tree video)
//	-This is when we have a primary index which is unique but we could also have a secondary index which is not unique
//Implement logic to create tables(I think these are just seperate b+ trees)
//Make sure we have critical db architecture set up
//	-Look at section 2 of the sqlite architecure and make sure we arent missing anything.
//Implement splitting algorithm in insert when page gets full

//TODO LATER:
//Implement LRU Cache or similart to update cache
//Implement journaling/write ahead logging(WAL) which saves data when db crashes
//data compression for query processing / b+ trees(claude mentioned LZ4, ZSTD, and delta encoding.)
