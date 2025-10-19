import { Language } from '../types/parser';
import { LessonData, SnippetData, indexedDBManager } from '../utils/indexedDB';

/**
 * ContentService manages lesson and snippet content for all supported languages
 */
export class ContentService {
  private static instance: ContentService;

  private constructor() {}

  public static getInstance(): ContentService {
    if (!ContentService.instance) {
      ContentService.instance = new ContentService();
    }
    return ContentService.instance;
  }

  /**
   * Initialize content service with default lessons and snippets
   */
  public async initialize(): Promise<void> {
    await this.loadDefaultContent();
  }

  /**
   * Get lessons for a specific language
   */
  public async getLessonsForLanguage(
    language: Language
  ): Promise<LessonData[]> {
    return await indexedDBManager.getLessonsByLanguage(language);
  }

  /**
   * Get snippets for a specific language
   */
  public async getSnippetsForLanguage(
    language: Language
  ): Promise<SnippetData[]> {
    return await indexedDBManager.getSnippetsByLanguage(language);
  }

  /**
   * Load default content for all languages
   */
  private async loadDefaultContent(): Promise<void> {
    // Load C++ content
    await this.loadCppContent();

    // Load Rust content
    await this.loadRustContent();
  }

  /**
   * Load C++ lessons and snippets
   */
  private async loadCppContent(): Promise<void> {
    const cppLessons: Omit<LessonData, 'cachedAt'>[] = [
      {
        id: 'cpp-intro-1',
        languageId: 'cpp',
        title: 'C++ Basics - Hello World',
        difficulty: 1,
        version: 1,
        content: `#include <iostream>

int main() {
    std::cout << "Hello, World!" << std::endl;
    return 0;
}`,
      },
      {
        id: 'cpp-intro-2',
        languageId: 'cpp',
        title: 'C++ Variables and Types',
        difficulty: 2,
        version: 1,
        content: `#include <iostream>

int main() {
    int age = 25;
    double height = 5.9;
    char grade = 'A';
    bool isStudent = true;
    
    std::cout << "Age: " << age << std::endl;
    std::cout << "Height: " << height << std::endl;
    std::cout << "Grade: " << grade << std::endl;
    std::cout << "Student: " << isStudent << std::endl;
    
    return 0;
}`,
      },
      {
        id: 'cpp-core-1',
        languageId: 'cpp',
        title: 'C++ Functions and Control Flow',
        difficulty: 3,
        version: 1,
        content: `#include <iostream>

int factorial(int n) {
    if (n <= 1) {
        return 1;
    }
    return n * factorial(n - 1);
}

int main() {
    for (int i = 1; i <= 5; ++i) {
        std::cout << "Factorial of " << i << " is " << factorial(i) << std::endl;
    }
    return 0;
}`,
      },
      {
        id: 'cpp-core-2',
        languageId: 'cpp',
        title: 'C++ Classes and Objects',
        difficulty: 4,
        version: 1,
        content: `#include <iostream>
#include <string>

class Person {
private:
    std::string name;
    int age;

public:
    Person(const std::string& n, int a) : name(n), age(a) {}
    
    void introduce() const {
        std::cout << "Hi, I'm " << name << " and I'm " << age << " years old." << std::endl;
    }
    
    void setAge(int newAge) {
        if (newAge >= 0) {
            age = newAge;
        }
    }
    
    int getAge() const {
        return age;
    }
};

int main() {
    Person person("Alice", 30);
    person.introduce();
    
    person.setAge(31);
    std::cout << "New age: " << person.getAge() << std::endl;
    
    return 0;
}`,
      },
      {
        id: 'cpp-advanced-1',
        languageId: 'cpp',
        title: 'C++ Templates and STL',
        difficulty: 5,
        version: 1,
        content: `#include <iostream>
#include <vector>
#include <algorithm>
#include <string>

template<typename T>
void printVector(const std::vector<T>& vec) {
    for (const auto& item : vec) {
        std::cout << item << " ";
    }
    std::cout << std::endl;
}

int main() {
    std::vector<int> numbers = {5, 2, 8, 1, 9, 3};
    std::vector<std::string> words = {"hello", "world", "cpp", "programming"};
    
    std::cout << "Original numbers: ";
    printVector(numbers);
    
    std::sort(numbers.begin(), numbers.end());
    std::cout << "Sorted numbers: ";
    printVector(numbers);
    
    std::cout << "Words: ";
    printVector(words);
    
    auto it = std::find(words.begin(), words.end(), "cpp");
    if (it != words.end()) {
        std::cout << "Found 'cpp' at position: " << std::distance(words.begin(), it) << std::endl;
    }
    
    return 0;
}`,
      },
    ];

    const cppSnippets: Omit<SnippetData, 'cachedAt'>[] = [
      {
        id: 'cpp-snippet-1',
        languageId: 'cpp',
        title: 'C++ Pointer Arithmetic',
        difficulty: 3,
        tags: ['pointers', 'memory'],
        sourceCode: `int arr[] = {10, 20, 30, 40, 50};
int* ptr = arr;

for (int i = 0; i < 5; ++i) {
    std::cout << "Value: " << *ptr << ", Address: " << ptr << std::endl;
    ++ptr;
}`,
      },
      {
        id: 'cpp-snippet-2',
        languageId: 'cpp',
        title: 'C++ Lambda Expressions',
        difficulty: 4,
        tags: ['lambda', 'functional'],
        sourceCode: `auto add = [](int a, int b) -> int {
    return a + b;
};

auto multiply = [](int x) {
    return [x](int y) { return x * y; };
};

std::cout << "Sum: " << add(5, 3) << std::endl;
auto multiplyBy5 = multiply(5);
std::cout << "5 * 7 = " << multiplyBy5(7) << std::endl;`,
      },
      {
        id: 'cpp-snippet-3',
        languageId: 'cpp',
        title: 'C++ Smart Pointers',
        difficulty: 5,
        tags: ['smart-pointers', 'memory-management'],
        sourceCode: `#include <memory>

class Resource {
public:
    Resource(int id) : id_(id) {
        std::cout << "Resource " << id_ << " created" << std::endl;
    }
    
    ~Resource() {
        std::cout << "Resource " << id_ << " destroyed" << std::endl;
    }
    
private:
    int id_;
};

std::unique_ptr<Resource> createResource(int id) {
    return std::make_unique<Resource>(id);
}

auto resource = createResource(42);`,
      },
    ];

    // Save lessons and snippets to IndexedDB
    for (const lesson of cppLessons) {
      await indexedDBManager.saveLesson(lesson as LessonData);
    }

    for (const snippet of cppSnippets) {
      await indexedDBManager.saveSnippet(snippet as SnippetData);
    }
  }

  /**
   * Load Rust lessons and snippets
   */
  private async loadRustContent(): Promise<void> {
    const rustLessons: Omit<LessonData, 'cachedAt'>[] = [
      {
        id: 'rust-intro-1',
        languageId: 'rust',
        title: 'Rust Basics - Hello World',
        difficulty: 1,
        version: 1,
        content: `fn main() {
    println!("Hello, World!");
}`,
      },
      {
        id: 'rust-intro-2',
        languageId: 'rust',
        title: 'Rust Variables and Mutability',
        difficulty: 2,
        version: 1,
        content: `fn main() {
    let x = 5;
    println!("The value of x is: {}", x);
    
    let mut y = 10;
    println!("The value of y is: {}", y);
    y = 15;
    println!("The value of y is now: {}", y);
    
    const MAX_POINTS: u32 = 100_000;
    println!("Maximum points: {}", MAX_POINTS);
}`,
      },
      {
        id: 'rust-core-1',
        languageId: 'rust',
        title: 'Rust Ownership and Borrowing',
        difficulty: 3,
        version: 1,
        content: `fn main() {
    let s1 = String::from("hello");
    let s2 = s1.clone();
    
    println!("s1 = {}, s2 = {}", s1, s2);
    
    let s3 = String::from("world");
    let len = calculate_length(&s3);
    
    println!("The length of '{}' is {}.", s3, len);
}

fn calculate_length(s: &String) -> usize {
    s.len()
}`,
      },
      {
        id: 'rust-core-2',
        languageId: 'rust',
        title: 'Rust Structs and Methods',
        difficulty: 4,
        version: 1,
        content: `#[derive(Debug)]
struct Rectangle {
    width: u32,
    height: u32,
}

impl Rectangle {
    fn new(width: u32, height: u32) -> Rectangle {
        Rectangle { width, height }
    }
    
    fn area(&self) -> u32 {
        self.width * self.height
    }
    
    fn can_hold(&self, other: &Rectangle) -> bool {
        self.width > other.width && self.height > other.height
    }
}

fn main() {
    let rect1 = Rectangle::new(30, 50);
    let rect2 = Rectangle::new(10, 40);
    
    println!("rect1 is {:?}", rect1);
    println!("The area of rect1 is {} square pixels.", rect1.area());
    println!("Can rect1 hold rect2? {}", rect1.can_hold(&rect2));
}`,
      },
      {
        id: 'rust-advanced-1',
        languageId: 'rust',
        title: 'Rust Enums and Pattern Matching',
        difficulty: 5,
        version: 1,
        content: `#[derive(Debug)]
enum Message {
    Quit,
    Move { x: i32, y: i32 },
    Write(String),
    ChangeColor(i32, i32, i32),
}

impl Message {
    fn call(&self) {
        match self {
            Message::Quit => println!("Quit message received"),
            Message::Move { x, y } => println!("Move to coordinates ({}, {})", x, y),
            Message::Write(text) => println!("Text message: {}", text),
            Message::ChangeColor(r, g, b) => println!("Change color to RGB({}, {}, {})", r, g, b),
        }
    }
}

fn main() {
    let messages = vec![
        Message::Quit,
        Message::Move { x: 10, y: 20 },
        Message::Write(String::from("Hello, Rust!")),
        Message::ChangeColor(255, 0, 0),
    ];
    
    for message in messages {
        message.call();
    }
    
    let some_number = Some(5);
    let some_string = Some("a string");
    let absent_number: Option<i32> = None;
    
    if let Some(value) = some_number {
        println!("Got a number: {}", value);
    }
}`,
      },
    ];

    const rustSnippets: Omit<SnippetData, 'cachedAt'>[] = [
      {
        id: 'rust-snippet-1',
        languageId: 'rust',
        title: 'Rust Iterator Patterns',
        difficulty: 3,
        tags: ['iterators', 'functional'],
        sourceCode: `let numbers = vec![1, 2, 3, 4, 5];

let doubled: Vec<i32> = numbers
    .iter()
    .map(|x| x * 2)
    .collect();

let sum: i32 = numbers
    .iter()
    .filter(|&&x| x % 2 == 0)
    .sum();

println!("Doubled: {:?}", doubled);
println!("Sum of evens: {}", sum);`,
      },
      {
        id: 'rust-snippet-2',
        languageId: 'rust',
        title: 'Rust Error Handling',
        difficulty: 4,
        tags: ['error-handling', 'result'],
        sourceCode: `use std::fs::File;
use std::io::ErrorKind;

fn read_file() -> Result<String, std::io::Error> {
    match File::open("hello.txt") {
        Ok(mut file) => {
            let mut contents = String::new();
            file.read_to_string(&mut contents)?;
            Ok(contents)
        }
        Err(error) => match error.kind() {
            ErrorKind::NotFound => {
                println!("File not found, creating new file");
                File::create("hello.txt")?;
                Ok(String::new())
            }
            other_error => Err(error),
        },
    }
}`,
      },
      {
        id: 'rust-snippet-3',
        languageId: 'rust',
        title: 'Rust Lifetimes and References',
        difficulty: 5,
        tags: ['lifetimes', 'references'],
        sourceCode: `fn longest<'a>(x: &'a str, y: &'a str) -> &'a str {
    if x.len() > y.len() {
        x
    } else {
        y
    }
}

struct ImportantExcerpt<'a> {
    part: &'a str,
}

impl<'a> ImportantExcerpt<'a> {
    fn level(&self) -> i32 {
        3
    }
    
    fn announce_and_return_part(&self, announcement: &str) -> &str {
        println!("Attention please: {}", announcement);
        self.part
    }
}

let string1 = String::from("abcd");
let string2 = "xyz";
let result = longest(string1.as_str(), string2);`,
      },
    ];

    // Save lessons and snippets to IndexedDB
    for (const lesson of rustLessons) {
      await indexedDBManager.saveLesson(lesson as LessonData);
    }

    for (const snippet of rustSnippets) {
      await indexedDBManager.saveSnippet(snippet as SnippetData);
    }
  }
}

// Export singleton instance
export const contentService = ContentService.getInstance();
