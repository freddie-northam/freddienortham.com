package content

// Role is a work history entry. This is structured data rather than markdown
// because it is a record, not writing: it has fixed fields and a fixed order.
type Role struct {
	Title   string
	Org     string
	URL     string
	Period  string
	Summary string
	Current bool
}

// Roles, newest first. The five most recent show on /projects/; the rest are
// behind "view all" so a co-founder role never sits in a list of ten and reads
// as a CV.
//
// The two automotive jobs (Fiskens, RM Sotheby's) are deliberately NOT here.
// They live under /not-code/cars/, where they explain the interest rather than
// padding the career.
var Roles = []Role{
	{
		Title: "Co-Founder & Head of Product", Org: "XALT", URL: "https://wearexalt.com/",
		Period: "Mar 2024 —", Current: true,
		Summary: "Product vision, platform strategy and innovation across XALT's ecosystem: tools and services powering brand–fan engagement, data intelligence and community growth.",
	},
	{
		Title: "Founder", Org: "Racinto",
		Period: "Jun 2025 —", Current: true,
		Summary: "A subsidiary of XALT building fan-first digital experiences across live events, media and entertainment. Clients include Motorsport Network.",
	},
	{
		Title: "Board Member", Org: "Y0LABS",
		Period: "Jun 2025 —", Current: true,
		Summary: "Transmedia storytelling. Its flagship, World of Montezuma, is the first playable sitcom in Fortnite.",
	},
	{
		Title: "Non-Executive Director", Org: "Northam Frederick",
		Period: "Jan 2024 —", Current: true,
		Summary: "Director and shareholder of an Edinburgh property firm converting listed buildings from commercial back to residential use.",
	},
	{
		Title: "Founder", Org: "SPANIEL Ltd.",
		Period: "Aug 2022 — Oct 2024",
		Summary: "A content moderation company protecting live broadcasts for Electronic Arts, Epic Games, Ubisoft, Valve, Twitch and Discord.",
	},
	{
		Title: "Advisor", Org: "Motorsport Network",
		Period: "Mar 2025 — Oct 2025",
		Summary: "Advised on digital fan engagement products including Race Center Live, covering platform innovation and creator-led community strategy.",
	},
	{
		Title: "Safety Lead", Org: "OS Studios",
		Period: "Dec 2021 — Sep 2023",
		Summary: "Brand safety guidelines and protocols for live and always-on projects. Clients included DoorDash Gaming, HelloFresh and eMLS.",
	},
	{
		Title: "Engagement & Safety Lead", Org: "Engine Shop",
		Period: "Feb 2020 — Jan 2022",
		Summary: "Led a team keeping fan spaces positive and inclusive for Major League Soccer, Bud Light and the Leukemia & Lymphoma Society.",
	},
}

// VisibleRoles is how many show before "view all".
const VisibleRoles = 5
