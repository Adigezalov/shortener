package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"time"
)

// ProfileCollector интерфейс для сбора профилей
type ProfileCollector interface {
	CollectHeapProfile(filename string) error
	CollectCPUProfile(filename string, duration time.Duration) error
	CollectGoroutineProfile(filename string) error
	CollectMutexProfile(filename string) error
}

// DefaultProfileCollector реализация ProfileCollector
type DefaultProfileCollector struct {
	profilesDir string
}

// NewProfileCollector создает новый сборщик профилей
func NewProfileCollector(profilesDir string) *DefaultProfileCollector {
	return &DefaultProfileCollector{
		profilesDir: profilesDir,
	}
}

// CollectHeapProfile собирает профиль использования памяти
func (pc *DefaultProfileCollector) CollectHeapProfile(filename string) error {
	// Принудительно запускаем GC для получения актуальной картины памяти
	runtime.GC()
	
	fullPath := filepath.Join(pc.profilesDir, filename)
	
	// Создаем директорию если она не существует
	if err := os.MkdirAll(pc.profilesDir, 0755); err != nil {
		return fmt.Errorf("failed to create profiles directory: %w", err)
	}
	
	f, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create heap profile file: %w", err)
	}
	defer f.Close()
	
	if err := pprof.WriteHeapProfile(f); err != nil {
		return fmt.Errorf("failed to write heap profile: %w", err)
	}
	
	return nil
}

// CollectCPUProfile собирает профиль использования CPU
func (pc *DefaultProfileCollector) CollectCPUProfile(filename string, duration time.Duration) error {
	fullPath := filepath.Join(pc.profilesDir, filename)
	
	// Создаем директорию если она не существует
	if err := os.MkdirAll(pc.profilesDir, 0755); err != nil {
		return fmt.Errorf("failed to create profiles directory: %w", err)
	}
	
	f, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create CPU profile file: %w", err)
	}
	defer f.Close()
	
	if err := pprof.StartCPUProfile(f); err != nil {
		return fmt.Errorf("failed to start CPU profile: %w", err)
	}
	
	// Ждем указанное время
	time.Sleep(duration)
	
	pprof.StopCPUProfile()
	return nil
}

// CollectGoroutineProfile собирает профиль горутин
func (pc *DefaultProfileCollector) CollectGoroutineProfile(filename string) error {
	fullPath := filepath.Join(pc.profilesDir, filename)
	
	// Создаем директорию если она не существует
	if err := os.MkdirAll(pc.profilesDir, 0755); err != nil {
		return fmt.Errorf("failed to create profiles directory: %w", err)
	}
	
	f, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create goroutine profile file: %w", err)
	}
	defer f.Close()
	
	profile := pprof.Lookup("goroutine")
	if profile == nil {
		return fmt.Errorf("goroutine profile not found")
	}
	
	if err := profile.WriteTo(f, 0); err != nil {
		return fmt.Errorf("failed to write goroutine profile: %w", err)
	}
	
	return nil
}

// CollectMutexProfile собирает профиль блокировок
func (pc *DefaultProfileCollector) CollectMutexProfile(filename string) error {
	fullPath := filepath.Join(pc.profilesDir, filename)
	
	// Создаем директорию если она не существует
	if err := os.MkdirAll(pc.profilesDir, 0755); err != nil {
		return fmt.Errorf("failed to create profiles directory: %w", err)
	}
	
	f, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create mutex profile file: %w", err)
	}
	defer f.Close()
	
	profile := pprof.Lookup("mutex")
	if profile == nil {
		return fmt.Errorf("mutex profile not found")
	}
	
	if err := profile.WriteTo(f, 0); err != nil {
		return fmt.Errorf("failed to write mutex profile: %w", err)
	}
	
	return nil
}

// CollectAllProfiles собирает все основные профили
func (pc *DefaultProfileCollector) CollectAllProfiles(baseFilename string) error {
	if err := pc.CollectHeapProfile(baseFilename + "_heap.pprof"); err != nil {
		return fmt.Errorf("failed to collect heap profile: %w", err)
	}
	
	if err := pc.CollectGoroutineProfile(baseFilename + "_goroutine.pprof"); err != nil {
		return fmt.Errorf("failed to collect goroutine profile: %w", err)
	}
	
	if err := pc.CollectMutexProfile(baseFilename + "_mutex.pprof"); err != nil {
		return fmt.Errorf("failed to collect mutex profile: %w", err)
	}
	
	return nil
}