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
	WPM          float64
	Accuracy     float64
	Duration     time.Duration
	Errors       int
	Mode         TestMode
	IsSuspicious bool
	CheatFlags   []string
}

type Player struct {
	Name            string
	BestWPM         float64
	TotalTests      int
	Level           int
	XP              int
	SuspiciousTests int
	CleanTests      int
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

	// Fraud detection indicator
	fraudPercent := 0.0
	if player.TotalTests > 0 {
		fraudPercent = float64(player.SuspiciousTests) / float64(player.TotalTests) * 100
	}

	var trustIndicator string
	if fraudPercent == 0 {
		trustIndicator = "✅ CLEAN"
	} else if fraudPercent < 20 {
		trustIndicator = "⚠️  CAUTION"
	} else {
		trustIndicator = "🚨 SUSPICIOUS"
	}

	fmt.Printf("👤 Player: %s | Level: %d | XP: %d/%d | Best WPM: %.1f | Tests: %d | Trust: %s\n",
		player.Name, level, xpProgress, xpForNext, player.BestWPM, player.TotalTests, trustIndicator)

	// XP Progress Bar
	barLength := 20
	progress := int(float64(xpProgress) / 100.0 * float64(barLength))
	bar := strings.Repeat("█", progress) + strings.Repeat("░", barLength-progress)
	fmt.Printf("XP Progress: [%s] %d%%\n", bar, xpProgress)

	if player.SuspiciousTests > 0 {
		fmt.Printf("🔍 Suspicious Tests: %d/%d (%.1f%%)\n", player.SuspiciousTests, player.TotalTests, fraudPercent)
	}
	fmt.Println()
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
	fmt.Println("🔍 Anti-cheat system is monitoring your performance...")
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

func detectFraud(original, typed string, duration time.Duration, wpm float64) (bool, []string) {
	var flags []string
	isSuspicious := false

	// 1. Unrealistic WPM Detection
	if wpm > 200 {
		flags = append(flags, "UNREALISTIC_SPEED")
		isSuspicious = true
	}

	// 2. Perfect Match Detection (Copy-Paste)
	if original == typed && duration.Seconds() < 10 {
		flags = append(flags, "INSTANT_PERFECT_MATCH")
		isSuspicious = true
	}

	// 3. Too Fast for Length
	expectedMinTime := float64(len(typed)) / 10.0 // Assume 10 chars per second max human speed
	if duration.Seconds() < expectedMinTime {
		flags = append(flags, "IMPOSSIBLY_FAST")
		isSuspicious = true
	}

	// 4. Consistent Speed Pattern (No human variation)
	words := strings.Fields(typed)
	if len(words) > 5 {
		avgTimePerWord := duration.Seconds() / float64(len(words))
		// Check if typing speed is too consistent (no natural variation)
		if avgTimePerWord < 0.2 { // Less than 0.2 seconds per word
			flags = append(flags, "ROBOTIC_CONSISTENCY")
			isSuspicious = true
		}
	}

	// 5. 100% Accuracy with High Speed
	accuracy := calculateAccuracy(original, typed)
	if accuracy == 100.0 && wpm > 80 {
		flags = append(flags, "PERFECT_HIGH_SPEED")
		isSuspicious = true
	}

	// 6. Exact Match with Minimal Time
	if strings.TrimSpace(original) == strings.TrimSpace(typed) && duration.Seconds() < 5 {
		flags = append(flags, "COPY_PASTE_DETECTED")
		isSuspicious = true
	}

	// 7. Burst Speed Detection (Too fast start)
	if wpm > 150 && duration.Seconds() < 30 {
		flags = append(flags, "BURST_SPEED_ANOMALY")
		isSuspicious = true
	}

	return isSuspicious, flags
}

func calculateAccuracy(original, typed string) float64 {
	if len(original) == 0 {
		return 0
	}

	correctChars := 0
	minLen := len(original)
	if len(typed) < minLen {
		minLen = len(typed)
	}

	for i := 0; i < minLen; i++ {
		if original[i] == typed[i] {
			correctChars++
		}
	}

	return float64(correctChars) / float64(len(original)) * 100
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

	// Fraud Detection
	isSuspicious, cheatFlags := detectFraud(original, typed, duration, wpm)

	return Result{
		WPM:          wpm,
		Accuracy:     accuracy,
		Duration:     duration,
		Errors:       errors,
		Mode:         mode,
		IsSuspicious: isSuspicious,
		CheatFlags:   cheatFlags,
	}
}

func displayResult(result Result, player *Player) {
	clearScreen()

	// Check for fraud first
	if result.IsSuspicious {
		fmt.Println("🚨 SUSPICIOUS ACTIVITY DETECTED! 🚨")
		fmt.Println("═══════════════════════════════════════")
		fmt.Println("⚠️  Our anti-cheat system has flagged this test as suspicious!")
		fmt.Println("🔍 Detected issues:")
		for _, flag := range result.CheatFlags {
			switch flag {
			case "UNREALISTIC_SPEED":
				fmt.Println("   • Typing speed exceeds human limits (>200 WPM)")
			case "INSTANT_PERFECT_MATCH":
				fmt.Println("   • Perfect text match completed too quickly")
			case "IMPOSSIBLY_FAST":
				fmt.Println("   • Completed faster than humanly possible")
			case "ROBOTIC_CONSISTENCY":
				fmt.Println("   • Typing pattern lacks human variation")
			case "PERFECT_HIGH_SPEED":
				fmt.Println("   • 100% accuracy at unrealistic speed")
			case "COPY_PASTE_DETECTED":
				fmt.Println("   • Evidence of copy-paste behavior")
			case "BURST_SPEED_ANOMALY":
				fmt.Println("   • Suspicious burst typing pattern")
			}
		}
		fmt.Println("\n🏆 This result will NOT count towards your records!")
		fmt.Println("💡 Tip: Type naturally for accurate results")
		player.SuspiciousTests++
	} else {
		fmt.Println("🎉 TEST COMPLETED! 🎉")
		fmt.Println("✅ Result verified as legitimate")
		player.CleanTests++
	}

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

	// Only award XP and records for clean tests
	if !result.IsSuspicious {
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
	} else {
		fmt.Println("\n❌ No XP or records awarded for suspicious tests")
		fmt.Println("🎮 Play fairly to earn rewards and track progress!")
	}

	player.TotalTests++

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

	// Fraud statistics
	if player.TotalTests > 0 {
		fmt.Printf("✅ Clean Tests: %d\n", player.CleanTests)
		fmt.Printf("🚨 Suspicious Tests: %d\n", player.SuspiciousTests)
		fraudPercent := float64(player.SuspiciousTests) / float64(player.TotalTests) * 100
		fmt.Printf("🔍 Trust Score: %.1f%% suspicious\n", fraudPercent)

		if fraudPercent == 0 {
			fmt.Println("🏅 Status: TRUSTED PLAYER")
		} else if fraudPercent < 20 {
			fmt.Println("⚠️  Status: CAUTION - Some suspicious activity")
		} else {
			fmt.Println("🚨 Status: HIGH SUSPICION - Multiple fraud flags")
		}
	}

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
	if player.SuspiciousTests == 0 && player.TotalTests >= 5 {
		fmt.Println("✅ Clean Player (No suspicious activity)")
	}

	fmt.Println("\nPress Enter to return to menu...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func main() {
	clearScreen()

	// Initialize player
	player := Player{
		Name:            "Typing Master",
		BestWPM:         0,
		TotalTests:      0,
		Level:           1,
		XP:              0,
		SuspiciousTests: 0,
		CleanTests:      0,
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
