package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type TestMode struct {
	Type     string
	Duration time.Duration
	Words    int
}

type Result struct {
	WPM      float64
	Accuracy float64
	Duration time.Duration
	Errors   int
	Mode     TestMode
}

type Player struct {
	Name       string
	BestWPM    float64
	TotalTests int
	Level      int
	XP         int
}

var wordBank = []string{
	"the", "of", "and", "a", "to", "in", "is", "you", "that", "it",
	"he", "was", "for", "on", "are", "as", "with", "his", "they", "i",
	"at", "be", "this", "have", "from", "or", "one", "had", "by", "word",
	"but", "not", "what", "all", "were", "we", "when", "your", "can", "said",
	"there", "each", "which", "she", "do", "how", "their", "time", "will", "about",
	"if", "up", "out", "many", "then", "them", "these", "so", "some", "her",
	"would", "make", "like", "into", "him", "has", "two", "more", "very", "after",
	"words", "first", "been", "who", "oil", "sit", "now", "find", "long", "down",
	"day", "did", "get", "come", "made", "may", "part", "over", "new", "sound",
	"take", "only", "little", "work", "know", "place", "year", "live", "me", "back",
	"give", "most", "very", "good", "woman", "through", "just", "form", "sentence", "great",
}

func clearScreen() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func printTitle() {
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    🚀 TYPING SPEED MASTER 🚀                  ║")
	fmt.Println("║              Test your typing skills and level up!           ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()
}

func printPlayerInfo(player Player) {
	level := player.Level
	xpForNext := (level + 1) * 100
	xpProgress := player.XP % 100

	fmt.Printf("👤 Player: %s | Level: %d | XP: %d/%d | Best WPM: %.1f | Tests: %d\n",
		player.Name, level, xpProgress, xpForNext, player.BestWPM, player.TotalTests)

	// XP Progress Bar
	barLength := 20
	progress := int(float64(xpProgress) / 100.0 * float64(barLength))
	bar := strings.Repeat("█", progress) + strings.Repeat("░", barLength-progress)
	fmt.Printf("XP Progress: [%s] %d%%\n\n", bar, xpProgress)
}

func showMenu() {
	fmt.Println("🎮 GAME MODES:")
	fmt.Println("1. ⏰ Time Mode (15s, 30s, 45s, 60s)")
	fmt.Println("2. 📝 Word Mode (15, 30, 45, 60 words)")
	fmt.Println("3. 🏆 View Stats")
	fmt.Println("4. 🚪 Exit")
	fmt.Print("\nSelect your challenge: ")
}

func selectTimeMode() TestMode {
	fmt.Println("\n⏰ TIME CHALLENGE")
	fmt.Println("1. 🔥 Quick Burst (15s)")
	fmt.Println("2. 💨 Speed Run (30s)")
	fmt.Println("3. 🎯 Focus Test (45s)")
	fmt.Println("4. 🚀 Endurance (60s)")
	fmt.Print("\nChoose your time limit: ")

	var choice int
	fmt.Scanln(&choice)

	durations := map[int]time.Duration{
		1: 15 * time.Second,
		2: 30 * time.Second,
		3: 45 * time.Second,
		4: 60 * time.Second,
	}

	if duration, ok := durations[choice]; ok {
		return TestMode{Type: "time", Duration: duration}
	}
	return TestMode{Type: "time", Duration: 30 * time.Second}
}

func selectWordMode() TestMode {
	fmt.Println("\n📝 WORD CHALLENGE")
	fmt.Println("1. 🔥 Sprint (15 words)")
	fmt.Println("2. 💨 Dash (30 words)")
	fmt.Println("3. 🎯 Marathon (45 words)")
	fmt.Println("4. 🚀 Ultra (60 words)")
	fmt.Print("\nChoose your word count: ")

	var choice int
	fmt.Scanln(&choice)

	words := map[int]int{
		1: 15,
		2: 30,
		3: 45,
		4: 60,
	}

	if wordCount, ok := words[choice]; ok {
		return TestMode{Type: "word", Words: wordCount}
	}
	return TestMode{Type: "word", Words: 30}
}

func generateText(wordCount int) string {
	rand.Seed(time.Now().UnixNano())
	words := make([]string, wordCount)

	for i := 0; i < wordCount; i++ {
		words[i] = wordBank[rand.Intn(len(wordBank))]
	}

	return strings.Join(words, " ")
}

func countdown() {
	for i := 3; i > 0; i-- {
		clearScreen()
		fmt.Printf("\n\n🚀 GET READY! Starting in %d...\n", i)
		time.Sleep(1 * time.Second)
	}
	clearScreen()
}

func runTest(mode TestMode) Result {
	var text string
	if mode.Type == "time" {
		text = generateText(200) // Generate enough words for time mode
	} else {
		text = generateText(mode.Words)
	}

	countdown()

	fmt.Println("💫 TYPE THE FOLLOWING TEXT:")
	fmt.Println("═══════════════════════════════════════")
	fmt.Println(text)
	fmt.Println("═══════════════════════════════════════")
	fmt.Println("🎯 Start typing now! (Press Enter when done)")
	fmt.Print("\n> ")

	startTime := time.Now()

	reader := bufio.NewReader(os.Stdin)
	var userInput string
	var endTime time.Time

	if mode.Type == "time" {
		// Time mode: read with timeout
		done := make(chan string)
		go func() {
			input, _ := reader.ReadString('\n')
			done <- strings.TrimSpace(input)
		}()

		select {
		case userInput = <-done:
			endTime = time.Now()
		case <-time.After(mode.Duration):
			endTime = time.Now()
			fmt.Println("\n⏰ Time's up!")
		}
	} else {
		// Word mode: read until done
		input, _ := reader.ReadString('\n')
		endTime = time.Now()
		userInput = strings.TrimSpace(input)
	}

	duration := endTime.Sub(startTime)
	return calculateResult(text, userInput, duration, mode)
}

func calculateResult(original, typed string, duration time.Duration, mode TestMode) Result {
	typedWords := strings.Fields(typed)

	// Calculate WPM
	minutes := duration.Minutes()
	wpm := float64(len(typedWords)) / minutes

	// Calculate accuracy
	correctChars := 0
	totalChars := len(original)
	errors := 0

	minLen := len(original)
	if len(typed) < minLen {
		minLen = len(typed)
	}

	for i := 0; i < minLen; i++ {
		if original[i] == typed[i] {
			correctChars++
		} else {
			errors++
		}
	}

	// Add errors for missing characters
	if len(typed) < len(original) {
		errors += len(original) - len(typed)
	} else if len(typed) > len(original) {
		errors += len(typed) - len(original)
	}

	accuracy := float64(correctChars) / float64(totalChars) * 100

	return Result{
		WPM:      wpm,
		Accuracy: accuracy,
		Duration: duration,
		Errors:   errors,
		Mode:     mode,
	}
}

func displayResult(result Result, player *Player) {
	clearScreen()

	fmt.Println("🎉 TEST COMPLETED! 🎉")
	fmt.Println("═══════════════════════════════════════")

	// Performance indicators
	var wpmRating string
	switch {
	case result.WPM >= 80:
		wpmRating = "🚀 LEGENDARY!"
	case result.WPM >= 60:
		wpmRating = "⭐ EXCELLENT!"
	case result.WPM >= 40:
		wpmRating = "💪 GOOD!"
	case result.WPM >= 20:
		wpmRating = "👍 AVERAGE"
	default:
		wpmRating = "🎯 KEEP PRACTICING!"
	}

	var accuracyRating string
	switch {
	case result.Accuracy >= 95:
		accuracyRating = "🎯 PERFECT!"
	case result.Accuracy >= 85:
		accuracyRating = "✨ GREAT!"
	case result.Accuracy >= 75:
		accuracyRating = "👌 GOOD!"
	default:
		accuracyRating = "📚 ROOM FOR IMPROVEMENT"
	}

	fmt.Printf("⚡ WPM: %.1f %s\n", result.WPM, wpmRating)
	fmt.Printf("🎯 Accuracy: %.1f%% %s\n", result.Accuracy, accuracyRating)
	fmt.Printf("⏱️  Duration: %.1fs\n", result.Duration.Seconds())
	fmt.Printf("❌ Errors: %d\n", result.Errors)

	// XP and leveling
	baseXP := int(result.WPM)
	bonusXP := 0
	if result.Accuracy >= 95 {
		bonusXP += 20
	} else if result.Accuracy >= 85 {
		bonusXP += 10
	}

	totalXP := baseXP + bonusXP
	player.XP += totalXP
	player.TotalTests++

	// Check for new personal best
	isNewBest := false
	if result.WPM > player.BestWPM {
		player.BestWPM = result.WPM
		isNewBest = true
		fmt.Println("🏆 NEW PERSONAL BEST! 🏆")
	}

	// Level up check
	newLevel := player.XP / 100
	if newLevel > player.Level {
		fmt.Printf("🌟 LEVEL UP! You're now level %d! 🌟\n", newLevel)
		player.Level = newLevel
	}

	fmt.Printf("\n💎 XP Earned: +%d", totalXP)
	if bonusXP > 0 {
		fmt.Printf(" (Base: %d + Accuracy Bonus: %d)", baseXP, bonusXP)
	}
	fmt.Println()

	if isNewBest {
		fmt.Println("🎊 Achievement Unlocked: New Speed Record!")
	}

	fmt.Println("\nPress Enter to continue...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func showStats(player Player) {
	clearScreen()
	fmt.Println("📊 YOUR TYPING STATISTICS 📊")
	fmt.Println("═══════════════════════════════════════")
	fmt.Printf("👤 Player: %s\n", player.Name)
	fmt.Printf("🏆 Best WPM: %.1f\n", player.BestWPM)
	fmt.Printf("🎮 Total Tests: %d\n", player.TotalTests)
	fmt.Printf("⭐ Current Level: %d\n", player.Level)
	fmt.Printf("💎 Total XP: %d\n", player.XP)

	// Level progress
	xpProgress := player.XP % 100
	fmt.Printf("📈 Progress to Level %d: %d/%d XP\n", player.Level+1, xpProgress, 100)

	// Achievements
	fmt.Println("\n🏅 ACHIEVEMENTS:")
	if player.BestWPM >= 20 {
		fmt.Println("✅ Speed Novice (20+ WPM)")
	}
	if player.BestWPM >= 40 {
		fmt.Println("✅ Typing Enthusiast (40+ WPM)")
	}
	if player.BestWPM >= 60 {
		fmt.Println("✅ Speed Demon (60+ WPM)")
	}
	if player.BestWPM >= 80 {
		fmt.Println("✅ Typing Master (80+ WPM)")
	}
	if player.TotalTests >= 10 {
		fmt.Println("✅ Dedicated Practitioner (10+ tests)")
	}
	if player.Level >= 5 {
		fmt.Println("✅ Rising Star (Level 5+)")
	}

	fmt.Println("\nPress Enter to return to menu...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func main() {
	clearScreen()

	// Initialize player
	player := Player{
		Name:       "Typing Master",
		BestWPM:    0,
		TotalTests: 0,
		Level:      1,
		XP:         0,
	}

	// Get player name
	fmt.Print("Enter your name: ")
	reader := bufio.NewReader(os.Stdin)
	name, _ := reader.ReadString('\n')
	player.Name = strings.TrimSpace(name)

	for {
		clearScreen()
		printTitle()
		printPlayerInfo(player)
		showMenu()

		var choice int
		fmt.Scanln(&choice)

		switch choice {
		case 1:
			mode := selectTimeMode()
			result := runTest(mode)
			displayResult(result, &player)

		case 2:
			mode := selectWordMode()
			result := runTest(mode)
			displayResult(result, &player)

		case 3:
			showStats(player)

		case 4:
			clearScreen()
			fmt.Println("🎮 Thanks for playing Typing Speed Master!")
			fmt.Printf("👋 See you later, %s! Keep practicing to improve your skills!\n", player.Name)
			return

		default:
			fmt.Println("❌ Invalid choice! Please try again.")
			time.Sleep(2 * time.Second)
		}
	}
}
