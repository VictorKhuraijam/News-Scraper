package templates

import "strings"

// Source represents a news source
type Source struct {
    ID   int
    Name string
}

// Sources is the list of available news sources
var Sources = []Source{
    {1, "TechCrunch"},
    {2, "BBC News"},
    {3, "The Guardian"},
    {5, "The Indian Express"},
}

func getCategoryClass(category string) string {
    switch category {
    case "technology":
        return "px-2 py-1 rounded-full text-xs font-medium bg-blue-100 text-blue-800"
    case "sports":
        return "px-2 py-1 rounded-full text-xs font-medium bg-green-100 text-green-800"
    case "politics":
        return "px-2 py-1 rounded-full text-xs font-medium bg-red-100 text-red-800"
    case "business":
        return "px-2 py-1 rounded-full text-xs font-medium bg-yellow-100 text-yellow-800"
    case "entertainment":
        return "px-2 py-1 rounded-full text-xs font-medium bg-purple-100 text-purple-800"
    case "health":
        return "px-2 py-1 rounded-full text-xs font-medium bg-pink-100 text-pink-800"
    default:
        return "px-2 py-1 rounded-full text-xs font-medium bg-gray-100 text-gray-800"
    }
}

func GetCategoryButtonClass(category string) string {
    switch category {
    case "technology":
        return "bg-blue-100 hover:bg-blue-200 text-blue-800 px-4 py-3 rounded-lg text-center font-medium transition"
    case "sports":
        return "bg-green-100 hover:bg-green-200 text-green-800 px-4 py-3 rounded-lg text-center font-medium transition"
    case "politics":
        return "bg-red-100 hover:bg-red-200 text-red-800 px-4 py-3 rounded-lg text-center font-medium transition"
    case "business":
        return "bg-yellow-100 hover:bg-yellow-200 text-yellow-800 px-4 py-3 rounded-lg text-center font-medium transition"
    case "entertainment":
        return "bg-purple-100 hover:bg-purple-200 text-purple-800 px-4 py-3 rounded-lg text-center font-medium transition"
    case "health":
        return "bg-pink-100 hover:bg-pink-200 text-pink-800 px-4 py-3 rounded-lg text-center font-medium transition"
    default:
        return "bg-gray-100 hover:bg-gray-200 text-gray-800 px-4 py-3 rounded-lg text-center font-medium transition"
    }
}

func Capitalize(s string) string {
    if s == "" {
        return s
    }
    return strings.ToUpper(s[:1]) + s[1:]
}
