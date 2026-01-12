package domsim

// CompareHTMLStructures 计算两个HTML body字符串之间的结构相似度。
// 返回值在 0.0 (完全不同) 到 1.0 (非常相似) 之间。
// 这是一个基础的实现，主要关注标签序列、节点数量和属性键的存在。
// func CompareHTMLStructures(htmlBody1 string, htmlBody2 string) float64 {
// 	// 快速检查完全相同或完全为空的情况
// 	if htmlBody1 == htmlBody2 {
// 		return 1.0
// 	}
// 	if len(htmlBody1) == 0 && len(htmlBody2) == 0 {
// 		return 1.0
// 	}
// 	// 如果一个为空而另一个不为空，则认为完全不相似
// 	if (len(htmlBody1) == 0 && len(htmlBody2) > 0) || (len(htmlBody1) > 0 && len(htmlBody2) == 0) {
// 		return 0.0
// 	}
//
// 	// 解析HTML Body 1
// 	parser1 := ast.Parser{}
// 	parser1.HttpParser(&htmlBody1) // HttpParser 会填充 parser1.tokenizer
// 	root1 := parser1.GetRoot()
// 	defer parser1.Clear() // 清理 parser1 内部状态
//
// 	// 解析HTML Body 2
// 	parser2 := ast.Parser{}
// 	parser2.HttpParser(&htmlBody2)
// 	root2 := parser2.GetRoot()
// 	defer parser2.Clear() // 清理 parser2 内部状态
//
// 	nodes1 := root1.Children
// 	nodes2 := root2.Children
//
// 	// 如果解析后都为空结构 (例如，只包含注释或无效HTML)，认为它们相似
// 	if len(nodes1) == 0 && len(nodes2) == 0 {
// 		return 1.0
// 	}
// 	// 如果解析后一个为空，另一个不为空，认为不相似
// 	if (len(nodes1) == 0 && len(nodes2) > 0) || (len(nodes1) > 0 && len(nodes2) == 0) {
// 		return 0.0
// 	}
//
// 	// 相似度评分的组成部分权重 (可调整)
// 	weightNodeCount := 0.2
// 	weightTagSequence := 0.5  // 标签序列匹配更重要
// 	weightAttributeSet := 0.3 // 属性集合匹配
//
// 	// 1. 节点数量相似度
// 	len1 := float64(len(nodes1))
// 	len2 := float64(len(nodes2))
// 	maxLen := math.Max(len1, len2)
// 	nodeCountScore := 0.0
// 	if maxLen > 0 { // 防止除以零
// 		nodeCountScore = 1.0 - (math.Abs(len1-len2) / maxLen)
// 	} else { // 如果 maxLen 是0 (即两个都是空节点列表，但前面已处理过，这里作为防御)
// 		nodeCountScore = 1.0
// 	}
//
// 	// 2. 标签序列和属性集相似度 (基于共同长度部分)
// 	commonLength := int(math.Min(len1, len2))
// 	tagSequenceMatchCount := 0.0
// 	avgAttributeSetScore := 0.0
//
// 	if commonLength > 0 {
// 		matchedTagsCountForAttrAvg := 0.0 // 记录实际参与属性比较的标签对数量
//
// 		for i := 0; i < commonLength; i++ {
// 			node1 := nodes1[i]
// 			node2 := nodes2[i]
//
// 			// 比较标签名 (忽略大小写)
// 			if strings.ToLower(node1.Value.TagName) == strings.ToLower(node2.Value.TagName) {
// 				tagSequenceMatchCount++
//
// 				// 3. 比较当前匹配标签的属性集合 (基于属性键的Jaccard相似度)
// 				attrs1Keys := make(map[string]bool)
// 				for _, attr := range node1.Value.Attributes {
// 					attrs1Keys[strings.ToLower(attr.Key)] = true
// 				}
// 				attrs2Keys := make(map[string]bool)
// 				for _, attr := range node2.Value.Attributes {
// 					attrs2Keys[strings.ToLower(attr.Key)] = true
// 				}
//
// 				intersectionSize := 0
// 				unionSize := len(attrs1Keys)
// 				// 计算并集大小，并将交集计数
// 				for key := range attrs2Keys {
// 					if attrs1Keys[key] {
// 						intersectionSize++
// 					} else {
// 						unionSize++ // 只在 key 不在 attrs1Keys 中时增加，避免重复计算交集部分
// 					}
// 				}
//
// 				if unionSize > 0 {
// 					avgAttributeSetScore += float64(intersectionSize) / float64(unionSize)
// 				} else if len(attrs1Keys) == 0 && len(attrs2Keys) == 0 { // 两者都没有属性
// 					avgAttributeSetScore += 1.0
// 				}
// 				// 如果一个有属性一个没有，Jaccard 自动为0 (因为 intersectionSize 为0，unionSize > 0)
// 				matchedTagsCountForAttrAvg++
// 			}
// 		}
//
// 		tagSequenceScore := tagSequenceMatchCount / float64(commonLength)
// 		attributeSetScore := 0.0
// 		if matchedTagsCountForAttrAvg > 0 {
// 			attributeSetScore = avgAttributeSetScore / matchedTagsCountForAttrAvg
// 		} else if commonLength > 0 && matchedTagsCountForAttrAvg == 0 {
// 			// 如果有共同长度的标签序列，但没有一对标签名是匹配的，那么属性得分应该是0
// 			attributeSetScore = 0.0
// 		} else { // commonLength == 0 或其他不太可能的情况
// 			attributeSetScore = 1.0 // 若无共同标签序列或无匹配标签，属性相似度默认为1(或0，取决于如何定义)
// 			// 设为1表示不因此项拉低总分；若严格，可设为0。
// 			// 鉴于前面有 nodeCountScore，这里设为1影响不大，或者可以基于commonLength为0时，此项不计入总分。
// 			// 考虑到前面 commonLength > 0 的条件，这里主要处理 matchedTagsCountForAttrAvg == 0 的情况。
// 			// 如果没有任何标签匹配，那么属性相似度可以认为是0，除非commonLength也为0。
// 			// 如果commonLength > 0 但没有标签匹配，则属性相似度为0是合理的。
// 			if commonLength > 0 { // 再次确认，如果commonLength > 0但无标签匹配
// 				attributeSetScore = 0.0
// 			} else { // commonLength == 0
// 				attributeSetScore = 1.0 // 如果没有共同序列，此项不应拉低分数
// 			}
// 		}
//
// 		// 最终加权平均
// 		finalScore := weightNodeCount*nodeCountScore +
// 			weightTagSequence*tagSequenceScore +
// 			weightAttributeSet*attributeSetScore
// 		return finalScore
//
// 	} else { // commonLength is 0 (一个或两个列表为空，或者长度都小于1但非空的情况不应该发生)
// 		// 如果两个都为空，前面已经返回1.0。如果一个为空一个不为空，也已返回0.0。
// 		// 此处是防御性的，如果 commonLength 是0，但 nodeCountScore 不是0或1（例如，都解析出非空但极短的列表）
// 		// 此时，我们只有节点数量得分。
// 		return weightNodeCount * nodeCountScore // 或者返回0，表示结构上几乎无共同点
// 	}
// }
//
// // CompareResponsesIntelligently 根据响应的Content-Type智能地选择比较方法。
// // contentType1, contentType2 应为小写。
// func CompareResponsesIntelligently(body1, contentType1, body2, contentType2 string) float64 {
// 	ct1Lower := strings.ToLower(contentType1)
// 	ct2Lower := strings.ToLower(contentType2)
//
// 	isHTML1 := strings.Contains(ct1Lower, "html") || strings.Contains(ct1Lower, "xhtml")
// 	isHTML2 := strings.Contains(ct2Lower, "html") || strings.Contains(ct2Lower, "xhtml")
//
// 	isJSON1 := strings.Contains(ct1Lower, "json")
// 	isJSON2 := strings.Contains(ct2Lower, "json")
//
// 	if isHTML1 && isHTML2 {
// 		return CompareHTMLStructures(body1, body2)
// 	} else if isJSON1 && isJSON2 {
// 		// 直接调用 strsim.Compare 进行字符串相似度比较
// 		return strsim.Compare(body1, body2)
// 	} else if body1 == body2 {
// 		return 1.0
// 	} else {
// 		if (isHTML1 && !isHTML2 && !isJSON2) || (isJSON1 && !isJSON2 && !isHTML2) || (isHTML1 && isJSON2) || (isJSON1 && isHTML2) {
// 			return 0.0 // 类型完全不兼容
// 		}
//
// 		// 如果类型模糊或都未知，则默认使用字符串比较作为回退
// 		return strsim.Compare(body1, body2)
// 	}
// }
