package openapiproject

import (
	"bufio"
	"context"
	"fmt"
	"regexp"
	"strings"
)

var (
	interfaceRx = regexp.MustCompile(`(?s)Interface:\s*([^\n]+)\n\s*Description:\s*([^\n]+)\s*Methods:\n(.*)`)
	methodRx    = regexp.MustCompile(`\s*-\s*([^\n]+)\s*Signature:\s*([^\n]+)\s*Description:\s*([^\n]+)`)
)

type llmResponseRaw struct {
	functionBody          string
	interfaces            string
	interfacesDescription string
}

type llmResponse struct {
	functionBody           string
	interfaces             string
	interfacesDescriptions []ProjectInterface
}

func ReadLLMResponse(body string) (*llmResponseRaw, error) {
	functionBody := strings.Builder{}
	interfaces := strings.Builder{}
	interfacesDescription := strings.Builder{}

	scanner := bufio.NewScanner(strings.NewReader(string(body)))
	state := 0
	for scanner.Scan() {
		if scanner.Err() != nil {
			return nil, scanner.Err()
		}
		line := scanner.Text()
		switch state {
		case 0:
			if strings.HasPrefix(line, "## function") {
				state = 1
			}
		case 1:
			if strings.HasPrefix(line, "## interfaces") {
				state = 2
			} else {
				functionBody.WriteString(line + "\n")
			}
		case 2:
			if strings.HasPrefix(line, "## interfaces_description") {
				state = 3
			} else {
				interfaces.WriteString(line + "\n")
			}
		case 3:
			interfacesDescription.WriteString(line + "\n")
		}
	}

	if state != 3 {
		return nil, fmt.Errorf("failed to parse body: invalid structure")
	}

	return &llmResponseRaw{
		functionBody:          functionBody.String(),
		interfaces:            interfaces.String(),
		interfacesDescription: interfacesDescription.String(),
	}, nil
}

func ParseInterfaceDescriptions(ifaceDesc string) []ProjectInterface {
	var descriptions []ProjectInterface
	spl := strings.SplitSeq(ifaceDesc, "```")
	for text := range spl {
		regexpMatches := interfaceRx.FindAllStringSubmatch(text, -1)
		projInterface := ProjectInterface{}
		for _, match := range regexpMatches {
			interfaceName := match[1]
			interfaceDescription := match[2]
			methodsText := match[3]
			projInterface.Name = interfaceName
			projInterface.Description = interfaceDescription
			pims := []ProjectInterfaceMethod{}
			methodMatches := methodRx.FindAllStringSubmatch(methodsText, -1)
			for _, methodMatch := range methodMatches {
				methodName := methodMatch[1]
				methodSignature := methodMatch[2]
				methodDescription := methodMatch[3]
				_, params, returns := ParseMethodSignature(methodSignature)
				pims = append(pims, ProjectInterfaceMethod{
					Name:        methodName,
					Description: methodDescription,
					Params:      params,
					Returns:     returns,
				})
			}
			projInterface.Methods = pims
			descriptions = append(descriptions, projInterface)
		}
	}

	return descriptions
}

func ParseMethodSignature(signature string) (name string, params []string, returns []string) {
	signature = strings.TrimSpace(signature)
	openParenIndex := strings.Index(signature, "(")
	closeParenIndex := strings.Index(signature, ")")
	if openParenIndex == -1 || closeParenIndex == -1 || closeParenIndex < openParenIndex {
		return "", nil, nil
	}
	name = strings.TrimSpace(signature[:openParenIndex])
	params = []string{}
	returns = []string{}
	paramsList := signature[openParenIndex+1 : closeParenIndex]
	returnsList := strings.TrimSpace(signature[closeParenIndex+1:])
	if len(paramsList) > 0 {
		params = strings.Split(paramsList, ",")
		for i := range params {
			params[i] = strings.TrimSpace(params[i])
		}
	}
	if len(returnsList) > 0 {
		returnsList = strings.Trim(strings.TrimSpace(returnsList), "()")
		returns = strings.Split(returnsList, ",")
		for i := range returns {
			returns[i] = strings.TrimSpace(returns[i])
		}
	}
	return
}

func GetMdBlock(body, blockName string) string {
	return strings.TrimSuffix(strings.TrimPrefix(body, "```"+blockName+"\n"), "```")
}

func (o *OpenAPIProject) ApplyResponse(ctx context.Context, content []byte, interfaces map[string]*ProjectInterface) error {
	// raw, err := ReadLLMResponse(string(content))
	// if err != nil {
	// 	return err
	// }

	// res := &llmResponse{
	// 	functionBody:           GetMdBlock(raw.functionBody, "go"),
	// 	interfaces:             GetMdBlock(raw.interfaces, "go"),
	// 	interfacesDescriptions: ParseInterfaceDescriptions(raw.interfacesDescription),
	// }

	return nil
}
